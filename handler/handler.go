package handler

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
)

//go:embed index.html
var html string

func Uplaod(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		io.WriteString(w, string(html))
	} else if r.Method == "POST" {
		uid := r.URL.Query().Get("uid")
		if uid == "" {
			sendMsg(w, 400, "权限不足")
			return
		}
		file, head, err := r.FormFile("file")
		if err != nil {
			sendMsg(w, 500, err.Error())
			return
		}
		defer file.Close()

		hash := md5.New()
		hash.Write([]byte(head.Filename))
		cipherText2 := hash.Sum(nil)
		hexText := make([]byte, 32)
		hex.Encode(hexText, cipherText2)

		

		name := uid + "-" + string(hexText) + path.Ext(head.Filename)

		newFile, err := os.Create(name)
		if err != nil {
			sendMsg(w, 500, err.Error())
			return
		}

		defer newFile.Close()

		_, err = io.Copy(newFile, file)

		if err != nil {
			sendMsg(w, 500, err.Error())
			return
		}

		sendMsg(w, 200, name)
	}
}

func sendMsg(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	resStr, _ := json.Marshal(struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}{Code: code, Msg: msg})

	io.WriteString(w, string(resStr))
}
