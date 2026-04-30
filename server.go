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
	passwords = make(map[string]string)
	commands  = make(map[string]string)
	mu        sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		pass := r.URL.Query().Get("pass")
		body, _ := ioutil.ReadAll(r.Body)
		
		mu.Lock()
		if len(body) > 0 { frames[id] = body }
		passwords[id] = pass
		cmd := commands[id]
		commands[id] = "" // Limpa o comando após o PC ler
		mu.Unlock()
		
		fmt.Fprint(w, cmd)
	})

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
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

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

	http.ListenAndServe(":"+port, nil)
}
