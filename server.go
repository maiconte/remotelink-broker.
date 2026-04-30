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

var (
	connections = make(map[string]*websocket.Conn)
	passwords   = make(map[string]string)
	mu          sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		role := r.URL.Query().Get("role")
		pass := r.URL.Query().Get("pass")
		
		mu.Lock()
		// Se for o PC, ele define a senha
		if role == "pc" {
			passwords[id] = pass
		} else {
			// Se for o Mobile, ele precisa acertar a senha
			if passwords[id] != pass {
				mu.Unlock()
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		mu.Unlock()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }

		mu.Lock()
		connections[id+role] = conn
		mu.Unlock()

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil { break }

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

	http.ListenAndServe(":"+port, nil)
}
