package main

import (
	"fmt"
	"net/http"
	"sync"
	"strings"
	"github.com/gorilla/websocket"
)

var (
	// Mapear conexões por ID
	pcClients     = make(map[string]*websocket.Conn)
	mobileClients = make(map[string]*websocket.Conn)
	mu            sync.Mutex
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "Standalone Broker Online") })

	// Rota para o PC (Envia Tela, Recebe Mouse)
	http.HandleFunc("/ws/pc", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(r.URL.Query().Get("id"))
		conn, _ := upgrader.Upgrade(w, r, nil)
		mu.Lock(); pcClients[id] = conn; mu.Unlock()
		fmt.Printf("💻 PC [%s] Conectado\n", id)
		
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil { break }
			// Se receber tela do PC, manda para o Celular correspondente
			mu.Lock()
			if mob, ok := mobileClients[id]; ok {
				mob.WriteMessage(mt, message)
			}
			mu.Unlock()
		}
	})

	// Rota para o Mobile (Recebe Tela, Envia Mouse)
	http.HandleFunc("/ws/mobile", func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToLower(r.URL.Query().Get("id"))
		conn, _ := upgrader.Upgrade(w, r, nil)
		mu.Lock(); mobileClients[id] = conn; mu.Unlock()
		fmt.Printf("📱 Mobile [%s] Conectado\n", id)

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil { break }
			// Se receber comando do Celular, manda para o PC
			mu.Lock()
			if pc, ok := pcClients[id]; ok {
				pc.WriteMessage(mt, message)
			}
			mu.Unlock()
		}
	})

	fmt.Println("🚀 Standalone Broker rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}
