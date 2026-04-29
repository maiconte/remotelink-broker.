package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var (
	frames = make(map[string][]byte)
	mu     sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	// PC envia imagem aqui
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		body, _ := ioutil.ReadAll(r.Body)
		if len(body) > 0 {
			mu.Lock()
			frames[id] = body
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	})

	// Mobile baixa imagem daqui
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "RemoteLink HTTP Engine ONLINE")
	})

	fmt.Println("🚀 Motor HTTP rodando na porta " + port)
	http.ListenAndServe(":"+port, nil)
}
