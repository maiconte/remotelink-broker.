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
		ReadBufferSize:  65536,
		WriteBufferSize: 65536,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Aceita TUDO
	}
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(r.URL.Query().Get("id"))
		role := r.URL.Query().Get("role")

		// LOG DE TENTATIVA
		fmt.Printf(">>> Tentativa de conexão: ID=%s ROLE=%s\n", id, role)

		if id == "" {
			fmt.Fprint(w, "Broker V3.2 ONLINE")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { 
			fmt.Println("❌ Erro no Upgrade:", err)
			return 
		}
		
		fullID := id + "_" + role
		mu.Lock()
		clients[fullID] = conn
		mu.Unlock()
		
		fmt.Printf("✅ CONECTADO: %s\n", fullID)

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
		fmt.Printf("❌ DESCONECTADO: %s\n", fullID)
	})

	fmt.Println("🚀 Servidor V3.2 na porta 8080")
	http.ListenAndServe(":8080", nil)
}
