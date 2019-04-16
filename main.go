package main

import (
	"net/http"
	"html/template"
	"io"
	"fmt"
	"strings"
	"os"
	"net/url"
)

func FormatTime(mil uint64) string {
	sec := mil / 1000
	return fmt.Sprintf("%d:%02d", sec / 60, sec % 60)
}

func main() {
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 {
			if strings.HasSuffix(r.URL.Path, ".png") {
				i := strings.Index(r.URL.Path[1:], "/") + 1
				http.ServeFile(w, r, "music" + r.URL.Path[:i] + "/sheet" + r.URL.Path[i:])
			} else {
				t, err := template.New("music.html").Funcs(template.FuncMap{
					"formattime": FormatTime,
					"div": func (a, b uint64) uint64 {
						return a / b
					},
					"url": url.Parse,
				}).ParseFiles("music.html")
				if err != nil {
					w.WriteHeader(500)
					io.WriteString(w, err.Error())
					return
				}
				piece, err := PieceFromId(r.URL.Path[1:])
				if err != nil {			
					w.WriteHeader(500)
					io.WriteString(w, err.Error())
					return
				}
				err = t.Execute(w, piece)
				if err != nil {
					io.WriteString(w, err.Error())
					return
				}
			}
		} else {
		    t, err := template.ParseFiles("index.html")
		    if err != nil {
		    	w.WriteHeader(500)
		    	io.WriteString(w, err.Error())
		    }
		    list, err := PieceList()
		    if err != nil {
		    	fmt.Fprintln(os.Stderr, err.Error())
		    }
		    if list == nil {
		    	fmt.Fprintln(os.Stderr, err)
		    }
		    if err = t.Execute(w, list); err != nil {
				io.WriteString(w, err.Error())
		    	return
		    }
		}
	})

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8123"
	}
	fmt.Fprintln(os.Stderr, http.ListenAndServe(":" + port, nil))
}
