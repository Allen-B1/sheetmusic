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
    os.Mkdir(".cache", 0777)

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 {
			if strings.HasSuffix(r.URL.Path, ".png") {
				i := strings.Index(r.URL.Path[1:], "/") + 1
				var imageid = r.URL.Path[i+1:len(r.URL.Path) - 4]
				var id = r.URL.Path[1:i]
				sheet, err := SheetFromId(id)
				if err != nil {
					w.WriteHeader(500)
					io.WriteString(w, err.Error())
					return
				}
				data, err := SheetRef(sheet[imageid]).Get()
				if err != nil {
					w.WriteHeader(500)
					io.WriteString(w, err.Error())
					return
				}	
				w.Header().Set("Content-Type", "image/png")
				w.Write(data)
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
		    if list == nil {
		    	fmt.Fprintln(os.Stderr, err)
		    }

			m := make(map[string][]*Piece)

			for _, piece := range list {
				if piece != nil {
					m[piece.Composer] = append(m[piece.Composer], piece)
				}
			}
		    
		    if err = t.Execute(w, m); err != nil {
				io.WriteString(w, err.Error())
		    	return
		    }
		}
	})

	http.HandleFunc("/style.css", func (w http.ResponseWriter, r* http.Request) {
		http.ServeFile(w, r, "style.css")		
	})

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8123"
	}
	fmt.Fprintln(os.Stderr, http.ListenAndServe(":" + port, nil))
}
