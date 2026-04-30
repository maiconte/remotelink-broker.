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
	commands = make(map[string]string)
	mu       sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// PC envia foto e recebe comando
	http.HandleFunc("/u", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		body, _ := ioutil.ReadAll(r.Body)
		mu.Lock()
		if len(body) > 0 { frames[id] = body }
		cmd := commands[id]
		commands[id] = ""
		mu.Unlock()
		fmt.Fprint(w, cmd)
	})

	// Celular baixa foto
	http.HandleFunc("/v", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		mu.Lock()
		img, ok := frames[id]
		mu.Unlock()
		if ok {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write(img)
		}
	})

	// Celular envia comando
	http.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		cmd := r.URL.Query().Get("cmd")
		mu.Lock()
		commands[id] = cmd
		mu.Unlock()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, "OK")
	})

	http.ListenAndServe(":"+port, nil)
}
