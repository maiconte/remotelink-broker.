package main

import (
	"fmt"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var (
	clients = make(map[string]*websocket.Conn)
	mu      sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// Middleware para habilitar CORS em todas as rotas
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	// Rota Raiz para a bolinha do App ficar Verde
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "RemoteLink Broker is Online 🚀")
	})

	// Rota WebSocket
	mux.HandleFunc("/ws/pc", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		mu.Lock()
		clients[id] = conn
		mu.Unlock()
		fmt.Printf("✅ PC [%s] conectado.\n", id)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				mu.Lock()
				delete(clients, id)
				mu.Unlock()
				break
			}
		}
	})

	// Rota de Sinal
	mux.HandleFunc("/api/signal", func(w http.ResponseWriter, r *http.Request) {
		targetID := r.URL.Query().Get("id")
		mu.Lock()
		conn, exists := clients[targetID]
		mu.Unlock()
		if exists {
			conn.WriteMessage(websocket.TextMessage, []byte("LAUNCH_ENGINE"))
			fmt.Fprint(w, "{\"status\":\"success\"}")
		} else {
			fmt.Fprint(w, "{\"status\":\"offline\"}")
		}
	})

	fmt.Println("🚀 Broker Global RemoteLink Pro rodando na porta 8080...")
	http.ListenAndServe(":8080", enableCORS(mux))
}
