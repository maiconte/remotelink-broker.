package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
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

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { return }
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "RemoteLink Broker is Online 🚀")
	})

	mux.HandleFunc("/ws/pc", func(w http.ResponseWriter, r *http.Request) {
		// Forçar ID para minúsculo
		id := strings.ToLower(r.URL.Query().Get("id"))
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		
		mu.Lock()
		clients[id] = conn
		mu.Unlock()
		
		fmt.Printf("✅ PC [%s] conectado.\n", id)
		
		conn.SetPongHandler(func(string) error { 
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil 
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				mu.Lock()
				delete(clients, id)
				mu.Unlock()
				fmt.Printf("❌ PC [%s] desconectado.\n", id)
				break
			}
		}
	})

	mux.HandleFunc("/api/signal", func(w http.ResponseWriter, r *http.Request) {
		// Forçar ID alvo para minúsculo
		targetID := strings.ToLower(r.URL.Query().Get("id"))
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

	fmt.Println("🚀 Broker Global RemoteLink Pro rodando...")
	http.ListenAndServe(":8080", enableCORS(mux))
}
