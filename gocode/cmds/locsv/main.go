package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/em4ht"
)

//go:embed frontend_build
var embeddedFiles3 embed.FS

func main() {
	var rootDir string
	var address string
	var noBrowser bool
	flag.StringVar(&rootDir, "root", "frontend_build", "path to the root of served files if live")
	flag.StringVar(&address, "address", ":3000", "host:port to listen to, defaults :3000")
	flag.BoolVar(&noBrowser, "no-browser", false, "prevents launching a browser window")
	flag.Parse()
	hp := strings.Split(address, ":")
	host := "localhost"
	port := "3000"
	if len(hp) == 2 {
		if len(hp[0]) > 0 {
			host = hp[0]
		}
		if len(hp[1]) > 0 {
			port = hp[1]
		}
	}

	fsys, err := em4ht.NewSpaFileSystem(embeddedFiles3, rootDir, "static", "/favicon.ico")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(fsys))
	http.HandleFunc("/api/", apiHandler3)
	if !noBrowser {
		go func() {
			time.Sleep(2000 * time.Millisecond)
			browser.OpenURL(fmt.Sprintf("http://%s:%s", host, port))
		}()
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}

func apiHandler3(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	fmt.Fprintf(w, "{\"version\": \"0.0.13.42\"}")
}
