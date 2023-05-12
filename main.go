package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cliclitv/clicli-cdn/handler"
)

func main() {
	http.HandleFunc("/upload", handler.Uplaod)
	http.Handle("/new_chunk", handleUploadChunk())
	http.Handle("/last_chunk", handleCompletedChunk())
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

func handleCompletedChunk() http.Handler {
	type request struct {
		UploadID string `json:"id"`
		Filename string `json:"name"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload request
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := handler.CompleteChunk(payload.UploadID, payload.Filename); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("file processed"))
	})
}
