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
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	// Rota Única para TUDO (PC e Mobile)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(r.URL.Query().Get("id"))
		role := r.URL.Query().Get("role")

		// Se não tiver ID, mostra status
		if id == "" {
			fmt.Fprint(w, "RemoteLink Broker is Online 🚀")
			return
		}

		// Se tiver ID, tenta virar WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		fullID := id + "_" + role
		mu.Lock()
		clients[fullID] = conn
		mu.Unlock()
		
		fmt.Printf("Conectado: %s\n", fullID)

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

	fmt.Println("🚀 Servidor Universal rodando...")
	http.ListenAndServe(":8080", nil)
}
