package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Mapa para conectar o PC ao Celular instantaneamente
var connections = make(map[string]*websocket.Conn)
var mu sync.Mutex

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// Rota Única e Universal para Pipeline (PC e Mobile)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		role := r.URL.Query().Get("role") // "pc" ou "mobile"
		
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }

		mu.Lock()
		connections[id+role] = conn
		mu.Unlock()

		fmt.Printf("🚀 %s conectado: %s\n", role, id)

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil { break }

			// Lógica de Repasse (Bridge)
			target := "mobile"
			if role == "mobile" { target = "pc" }

			mu.Lock()
			if targetConn, ok := connections[id+target]; ok {
				targetConn.WriteMessage(msgType, msg)
			}
			mu.Unlock()
		}

		mu.Lock()
		delete(connections, id+role)
		mu.Unlock()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "V5 ENGINE X-TREME ONLINE") })
	
	fmt.Println("🔥 Motor Definitivo rodando na porta " + port)
	http.ListenAndServe(":"+port, nil)
}
