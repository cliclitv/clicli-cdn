package main

import (
	"fmt"
	"net/http"

	"github.com/cliclitv/clicli-cdn/handler"
)

func main() {
	http.HandleFunc("/upload", handler.Uplaod)
	err:=http.ListenAndServe(":2333", nil)
	if err!= nil{
		fmt.Println(err.Error())
	}
}
