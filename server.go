package main

import (
	"fmt"
	"net/http"
	"sync"
	"strings"
	"os"
	"github.com/gorilla/websocket"
)

var (
	clients = make(map[string]*websocket.Conn)
	mu      sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// Rota para o PC
	http.HandleFunc("/pc/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/pc/"))
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		mu.Lock(); clients[id+"_pc"] = conn; mu.Unlock()
		fmt.Printf("💻 PC [%s] Conectado\n", id)

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil { break }
			mu.Lock()
			if mob, ok := clients[id+"_mobile"]; ok { mob.WriteMessage(mt, message) }
			mu.Unlock()
		}
		mu.Lock(); delete(clients, id+"_pc"); mu.Unlock()
	})

	// Rota para o Mobile
	http.HandleFunc("/mobile/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/mobile/"))
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		mu.Lock(); clients[id+"_mobile"] = conn; mu.Unlock()
		fmt.Printf("📱 Mobile [%s] Conectado\n", id)

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil { break }
			mu.Lock()
			if pc, ok := clients[id+"_pc"]; ok { pc.WriteMessage(mt, message) }
			mu.Unlock()
		}
		mu.Lock(); delete(clients, id+"_mobile"); mu.Unlock()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "V3.4 Active") })

	fmt.Println("🚀 Servidor V3.4 na porta " + port)
	http.ListenAndServe(":"+port, nil)
}
