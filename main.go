package main

import (
	"fmt"
	"net/http"
	"github.com/cliclitv/clicli-cdn/handler"
)

func main() {
	http.HandleFunc("/last_chunk", handler.Uplaod)
	http.Handle("/new_chunk", handleUploadChunk())
	err := http.ListenAndServe(":2333", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func handleUploadChunk() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler.ProcessChunk(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("chunk processed"))
	})
}