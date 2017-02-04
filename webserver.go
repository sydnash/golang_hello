package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io"
	_ "io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	_ "runtime/debug"
	"strconv"
	"time"
)

var addr = flag.String("addr", ":1718", "http service address") // Q=17, R=18

var installTemplate = `
<!DOCTYPE html>
<html>
<head lang="en">
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1, minimum-scale=1, user-scalable=no">
  <meta content="telephone=no" name="format-detection"/>
  <title>application name</title>
</head>
<style>
  html {
    width: 100%;
    height: 100%
  }
  body {
    width: 100%;
    height: 100%;
    background-color: #fafafa;
    font-family: "Microsoft YaHei";
    color: #0a0909;
    -webkit-touch-callout: none;
    -webkit-user-select: none;
  }
  div, p, header, footer, h1, h2, h3, h4, h5, h6, span, i, b, em, ul, li, dl, dt, dd, body, input, select, form, button {
    margin: 0;
    padding: 0
  }
  ul, li {
    list-style: none
  }
  img {
    border: 0 none
  }
  input, img {
    vertical-align: middle
  }
  * {
    -webkit-tap-highlight-color: rgba(0, 0, 0, 0);
    outline: 0;
    box-sizing: border-box;
  }
  a {
    text-decoration: none;
  }
  h1, h2, h3, h4, h5, h6 {
    font-weight: normal;
  }
  body {
    padding: 20px;
  }
  .title {
    font-size: 18px;
    margin-bottom: 20px;
  }
  .install {
    width: 150px;
    height: 40px;
    border: 1px solid #ccc;
    background: transparent;
    border-radius: 6px;
    font-size: 14px;
    margin-bottom: 10px;
    display: block;
  }
</style>
<body>
  <p class="title">iOS应用OTA安装</p>
  <a href="itms-services://?action=download-manifest&url=https://192.168.23.7:1718/static/manifest.plist">
    <button class="install">安装应用</button>
  </a>
  <a title="iPhone" href="http://192.168.23.7:1717/static/ca.crt">
    <button class="install">证书信任</button>
  </a>
</body>
</html>
`

var uploadStr = `
  <html>
    <head>
        <title>上传文件</title>
    </head>
    <body>
    <form enctype="multipart/form-data" action="/upload" method="post">
      <input type="file" name="uploadfile" />
      <input type="hidden" name="token" value="{{.}}"/>
      <input type="submit" value="upload" />
    </form>
    </body>
    </html>
`
var templ = template.Must(template.New("qr").Parse(templateStr))

var g_token string

// 处理/upload 逻辑
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //获取请求的方法
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		g_token = token
		t, _ := template.New("upload").Parse(uploadStr)
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)

		token := r.FormValue("token")
		if token == g_token {
			g_token = ""
		} else {
			http.Redirect(w, r, "upload", http.StatusMovedPermanently)
			return
		}

		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	http.Handle("/", http.HandlerFunc(QR))
	http.Handle("/install/", http.HandlerFunc(install))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("test"))))
	http.HandleFunc("/upload", upload)
	fmt.Println(*addr)
	go func() {
		http.ListenAndServe(":1717", nil)
	}()
	err := http.ListenAndServeTLS(*addr, "server.crt", "server.key", nil)
	//err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
func sessionId() string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
func install(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, installTemplate)
}
func QR(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(string(debug.Stack()))
	cookie := &http.Cookie{Name: sessionId(), Value: "t"}
	fmt.Println("random sessionid:", cookie.Name)
	http.SetCookie(w, cookie)
	templ.Execute(w, req.FormValue("s"))
}

const templateStr = `
<html>
    <head>
        <title>QR Link Generator</title>
    </head>
    <body>
        {{if .}}
                <img src="http://chart.apis.google.com/chart?chs=300x300&cht=qr&choe=UTF-8&chl={{.}}" />
                <br>
                    {{.}}
                <br>
            <br>
        {{end}}
        <form action="/" name=f method="GET">
            <input maxLength=1024 size=70 name=s value="" title="Text to QR Encode">
            <input type=submit value="Show QR" name=qr>
        </form>
    </body>
</html>
`
