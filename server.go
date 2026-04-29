package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var (
	frames   = make(map[string][]byte)
	commands = make(map[string]string) // Guarda o último comando do mouse
	mu       sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// PC envia imagem e RECEBE comandos
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		body, _ := ioutil.ReadAll(r.Body)
		if len(body) > 0 {
			mu.Lock()
			frames[id] = body
			mu.Unlock()
		}

		// Retorna o comando pendente para o PC executar
		mu.Lock()
		cmd := commands[id]
		commands[id] = "" // Limpa após ler
		mu.Unlock()
		
		fmt.Fprint(w, cmd)
	})

	// Celular envia comando de clique
	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		cmd := r.URL.Query().Get("cmd")
		
		mu.Lock()
		commands[id] = cmd
		mu.Unlock()
		
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, "OK")
	})

	// Mobile baixa imagem
	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		mu.Lock()
		img, ok := frames[id]
		mu.Unlock()

		if ok {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write(img)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "V3.5 Control Ready") })

	fmt.Println("🚀 Motor de Controle V3.5 na porta " + port)
	http.ListenAndServe(":"+port, nil)
}
