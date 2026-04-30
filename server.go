package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var (
	frames    = make(map[string][]byte)
	passwords = make(map[string]string) // Guarda a senha de cada ID
	commands  = make(map[string]string)
	mu        sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// PC envia imagem e DEFINE a senha
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		pass := r.URL.Query().Get("pass")
		body, _ := ioutil.ReadAll(r.Body)
		
		mu.Lock()
		if len(body) > 0 { frames[id] = body }
		passwords[id] = pass // Atualiza a senha atual do PC
		cmd := commands[id]
		commands[id] = ""
		mu.Unlock()
		
		fmt.Fprint(w, cmd)
	})

	// Mobile só recebe se a SENHA estiver correta
	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		pass := r.URL.Query().Get("pass")
		
		mu.Lock()
		correctPass := passwords[id]
		img, ok := frames[id]
		mu.Unlock()

		if ok && pass == correctPass {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write(img)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusUnauthorized) // 401 - Senha Errada
		}
	})

	// Controle também exige senha
	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		pass := r.URL.Query().Get("pass")
		cmd := r.URL.Query().Get("cmd")
		
		mu.Lock()
		if pass == passwords[id] { commands[id] = cmd }
		mu.Unlock()
		
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, "OK")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "RemoteLink Ultra Secure") })

	http.ListenAndServe(":"+port, nil)
}
