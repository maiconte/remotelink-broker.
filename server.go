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
	// Pega a porta do Render ou usa 8080 como reserva
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(r.URL.Query().Get("id"))
		role := r.URL.Query().Get("role")

		if id == "" {
			fmt.Fprint(w, "RemoteLink V3.3 Active")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		fullID := id + "_" + role
		mu.Lock()
		clients[fullID] = conn
		mu.Unlock()
		
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil { break }
			
			targetRole := "mobile"
			if role == "mobile" { targetRole = "pc" }
			
			mu.Lock()
			if target, ok := clients[id+"_"+targetRole]; ok {
				target.WriteMessage(mt, message)
			}
			mu.Unlock()
		}
		mu.Lock(); delete(clients, fullID); mu.Unlock()
	})

	fmt.Println("🚀 Servidor V3.3 rodando na porta " + port)
	http.ListenAndServe(":"+port, nil)
}
