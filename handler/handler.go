package handler

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"
	"os"
	"path"
	"fmt"
	"path/filepath"
	"os/exec"
	"strings"
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

		folderPath:=createDateDir("")

		name := folderPath+"/"+uid + "-" + string(hexText) + path.Ext(head.Filename)

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

		go Transform(name)


		sendMsg(w, 200, name)
	}
}

func Transform(name string){
		// 设置 ffmpeg 命令行参数
		dir:= strings.Replace(name,".mp4","/",1)
		i:="./"+dir+"index.m3u8"
		o:="./"+dir+"out%03d.ts"
		createDateDir(dir)

		args := []string{"-i",
		"./"+name,
		"-c",
		"copy",
		"-vbsf",
		"h264_mp4toannexb",
		"-map",
		"0",
		"-f",
		"segment",
		"-segment_list",
		i,
		"-segment_time",
		"6",
		o}

		// 创建 *exec.Cmd
		cmd := exec.Command("ffmpeg", args...)
	
		// 运行 ffmpeg 命令
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			return
		}
}

// https://github.com/cliclitv/clicli-cdn.git

func createDateDir(path string)string{
	var folderName string
	if path == ""{
		folderName=time.Now().Format("20060102")
	}else{
		folderName = path
	}
	
	folderPath := filepath.Join("./", folderName)

	
	_,err:= os.Stat(folderPath); 
	
	if os.IsNotExist(err){
		os.Mkdir(folderPath,0777)
		os.Chmod(folderPath,0777)
	}
	return folderPath
}

func sendMsg(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	resStr, _ := json.Marshal(struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}{Code: code, Msg: msg})

	io.WriteString(w, string(resStr))
}
