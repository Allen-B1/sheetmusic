package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sort"
	"encoding/json"
	"errors"
	"time"
	"strings"
)

type Piece struct {
	Id string // id of piece
	Name string // name
	Audio string // url to audio
	Artist string
	Description string
	Composer string
	Color string
	Map map[uint64]string // map
	MovementTimes map[string]uint64
	MovementList []string
	SheetCredits map[string]string
}

func ToString(val interface{}) string {
	if (val == nil) {
		return ""
	}
	return fmt.Sprint(val)
}

func PieceList() ([]*Piece, error) {	
	files, err := ioutil.ReadDir("./music")
	if err != nil {
		return nil, err
	}

	out := make([]*Piece, len(files))
	for i, file := range files {
		out[i], err = PieceFromId(file.Name())
	}
	return out, err
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
	am := make(map[uint64]string)
	for str, page := range amraw {
		timestamp, err := time.ParseDuration(str)
		if err != nil {
			return nil, err
		}
		am[uint64(timestamp) / 1000000] = fmt.Sprint(page)
	}

	out := &Piece{
		Id: id,
		Name: ToString(m["name"]),
		Audio: ToString(m["audio"]),
		Artist: ToString(m["audio_artist"]),
		Description: ToString(m["description"]),
		Composer: ToString(m["composer"]),
		Color: ToString(m["color"]),
		Map: am,
	}

	if m["mvmts"] != nil {	
		out.MovementTimes = make(map[string]uint64)
	
		mvmts, ok := m["mvmts"].(map[string]interface{})
		if !ok {
			return nil, errors.New("Key \"mvmts\" is wrong type")
		}

		for mvmt, timeInter := range mvmts {
			timeStr, ok := timeInter.(string)
			if !ok { return nil, errors.New("Key \"mvmts\" is wrong type") }

			time, err := time.ParseDuration(timeStr)
			if err != nil {
				return nil, err
			}

			dot := strings.Index(mvmt, ".")
			
			out.MovementTimes[mvmt[dot+2:]] = uint64(time) / 1000000
			out.MovementList = append(out.MovementList, mvmt)
		}

		sort.Slice(out.MovementList, func (i, j int) bool {
			iDot := strings.Index(out.MovementList[i], ".")
			jDot := strings.Index(out.MovementList[j], ".")
			if iDot < 0 || jDot < 0 {
				return false
			}

			iNum, _ := strconv.ParseUint(out.MovementList[i][:iDot], 10, 16)
			jNum, _ := strconv.ParseUint(out.MovementList[j][:jDot], 10, 16)

			return iNum < jNum
		})

		for i, mvmt := range out.MovementList {
			dot := strings.Index(mvmt, ".")
			out.MovementList[i] = mvmt[dot+2:]
		}
	}
 	
	return out, nil
}