package handler

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//go:embed index.html
var html string

func Uplaod(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if r.Method == "GET" {
		io.WriteString(w, string(html))
	} else if r.Method == "POST" {
		uid := r.URL.Query().Get("uid")
		if uid == "" {
			sendMsg(w, 400, "权限不足")
			return
		}

		type request struct {
			UploadID string `json:"id"`
			Filename string `json:"name"`
		}

		var payload request
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			sendMsg(w, 500, err.Error())
			return
		}

		hash := md5.New()
		hash.Write([]byte(payload.Filename))
		cipherText2 := hash.Sum(nil)
		hexText := make([]byte, 32)
		hex.Encode(hexText, cipherText2)

		folderPath := CreateDateDir("")

		name := folderPath + "/" + uid + "_" + string(hexText) + path.Ext(payload.Filename)

		// 先合并

		if err := CompleteChunk(payload.UploadID, payload.Filename); err != nil {
			sendMsg(w, 500, err.Error())
			return 
		}

		go Transform(name)

		sendMsg(w, 200, name)
	}
}

func Transform(name string) {
	// 设置 ffmpeg 命令行参数
	dir := strings.Replace(name, ".mp4", "/", 1)
	i := "./" + dir + "index.m3u8"
	o := "./" + dir + "out%03d.ts"
	CreateDateDir(dir)

	//libfdk_aac
	args := []string{
		"-i",
		"./" + name,
		"-i", "logo.png",
		"-filter_complex", `[1][0]scale2ref=iw/8:iw/16 [b][a];[a][b] overlay=10:10,scale=1920:-2,pad=iw:1080:0:(oh-ih)/2:black`,
		"-c:v", "libx264", "-b:v", "2000k", "-c:a", "copy",
		// "-vf", `movie=logo.png, scale=` + output + `*0.4:-1 [logo]; [in][logo] overlay=10:10, scale=1920:-2,pad=iw:1080:0:(oh-ih)/2:black [out]`,
		"-map",
		"0",
		"-f",
		"segment",
		"-segment_list",
		i,
		"-segment_time",
		"3",
		o}

	// 创建 *exec.Cmd
	cmd := exec.Command("ffmpeg", args...)

	fmt.Println(cmd)

	// 运行 ffmpeg 命令
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return
	}

	// 最后删除文件

	_ = os.Remove(name)
}

func CreateDateDir(path string) string {
	var folderName string
	if path == "" {
		folderName = time.Now().Format("20060102")
	} else {
		folderName = path
	}

	folderPath := filepath.Join("./", folderName)

	_, err := os.Stat(folderPath)

	if os.IsNotExist(err) {
		os.Mkdir(folderPath, 0777)
		os.Chmod(folderPath, 0777)
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

func Cmd(commandName string, params []string) (string, error) {
	cmd := exec.Command(commandName, params...)
	fmt.Println("Cmd", cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	return out.String(), err
}
