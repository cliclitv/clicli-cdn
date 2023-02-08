package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func Uplaod(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := ioutil.ReadFile("./static/upload.html")
		if err != nil {
			sendMsg(w, 500, err.Error())
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		pid := r.URL.Query().Get("pid")
		oid := r.URL.Query().Get("oid")
		file, head, err := r.FormFile("file")
		if err != nil {
			sendMsg(w, 500, err.Error())
			return
		}
		defer file.Close()

		newFile, err := os.Create("temp-" + pid + "-" + oid + head.Filename)
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

		sendMsg(w, 200, "上传成功")
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
