package main

import (
	"net/http"
	"html/template"
	"io"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"errors"
	"time"
	"strings"
)

type Piece struct {
	Id string // id of piece
	Name string // name
	Audio string // url to audio
	Map map[uint64]uint // map
}

func PieceFromId(id string) (*Piece, error) {
	body, err := ioutil.ReadFile("music/" + id + "/data.json")
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}

	amraw, ok := m["map"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Key \"map\" is not present or is the wrong type")
	}
	am := make(map[uint64]uint)
	for str, page := range amraw {
		timestamp, err := time.ParseDuration(str)
		if err != nil {
			return nil, err
		}
		pagef, ok := page.(float64)
		if !ok {
			return nil, errors.New("\"map\"[\"" + str + "\"] is not an integer")
		}
		am[uint64(timestamp) / 1000000] = uint(pagef)
	}
 	
	return &Piece{
		Id: id,
		Name: fmt.Sprint(m["name"]),
		Audio: fmt.Sprint(m["audio"]),
		Map: am,
	}, nil
}

func main() {
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 {
			if strings.HasSuffix(r.URL.Path, ".png") {
				i := strings.Index(r.URL.Path[1:], "/") + 1
				http.ServeFile(w, r, "music" + r.URL.Path[:i] + "/sheet" + r.URL.Path[i:])
			} else {
				t, err := template.ParseFiles("music.html")
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
		    http.ServeFile(w, r, "home.html")
		}
	})
	http.ListenAndServe(":8123", nil)
}
