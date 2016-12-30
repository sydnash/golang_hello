package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":1718", "http service address") // Q=17, R=18

var tmp = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<title>应用名字</title>
</head>
<body>
<h1 style="font-size:40pt">iOS应用OTA安装<h1/>
<h1 style="font-size:40pt">
<a href="itms-services://?action=download-manifest&url=https://192.168.23.7:1718/static/manifest.plist">Install App</a>
<h1/>
<a title="iPhone" href="http://192.168.23.7:1717/static/ca.crt">ssl 证书安装</a>
<h1/>
</body>
</html>
`
var templ = template.Must(template.New("qr").Parse(templateStr))

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	http.Handle("/", http.HandlerFunc(QR))
	http.Handle("/install/", http.HandlerFunc(install))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("test"))))
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
	fmt.Fprint(w, tmp)
}
func QR(w http.ResponseWriter, req *http.Request) {
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
