package main

import (
	"fmt"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var (
	// Gerenciador de conexões ativas (ID -> Conexão)
	clients = make(map[string]*websocket.Conn)
	mu      sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	// Endpoint WebSocket para os computadores (PCs) ficarem conectados
	http.HandleFunc("/ws/pc", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		mu.Lock()
		clients[id] = conn
		mu.Unlock()
		
		fmt.Printf("✅ PC [%s] conectado via Cloud Link.\n", id)
		
		// Manter a conexão viva
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				mu.Lock()
				delete(clients, id)
				mu.Unlock()
				break
			}
		}
	})

	// Endpoint para o Celular mandar o comando de conexão
	http.HandleFunc("/api/signal", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		targetID := r.URL.Query().Get("id")
		
		mu.Lock()
		conn, exists := clients[targetID]
		mu.Unlock()
		
		if exists {
			// Enviar comando para o PC alvo abrir o motor nativo
			conn.WriteMessage(websocket.TextMessage, []byte("LAUNCH_ENGINE"))
			fmt.Printf("📡 Comando enviado para o PC [%s]\n", targetID)
			fmt.Fprint(w, "{\"status\":\"success\"}")
		} else {
			fmt.Fprint(w, "{\"status\":\"offline\"}")
		}
	})

	fmt.Println("🚀 Broker Global RemoteLink Pro rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}
