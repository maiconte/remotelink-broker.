package main

import (
	"fmt"
	"net/http"
	"sync"
	"strings"
	"github.com/gorilla/websocket"
)

var (
	clients = make(map[string]*websocket.Conn)
	mu      sync.Mutex
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024 * 1024,
		WriteBufferSize: 1024 * 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "RemoteLink Engine V3 Online") })

	http.HandleFunc("/ws/stream", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(r.URL.Query().Get("id"))
		role := r.URL.Query().Get("role") // "pc" ou "mobile"
		
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		fullID := id + "_" + role
		mu.Lock()
		clients[fullID] = conn
		mu.Unlock()
		
		fmt.Printf("Connected: %s as %s\n", id, role)

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil { break }
			
			// Roteamento Direto
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

	fmt.Println("🚀 Servidor V3 Ativo na porta 8080")
	http.ListenAndServe(":8080", nil)
}
