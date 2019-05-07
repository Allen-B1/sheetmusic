package main

import (
    "strings"
	"net/http"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"errors"
	"io/ioutil"
	"regexp"
	"bytes"
	"math/rand"
)

// Maps url to file path
var sheetRefCache = make(map[string]string)

type SheetRef []string

func pdfHeight(path string, page uint) uint {
	out, err := exec.Command("pdfinfo", path,
		"-f", fmt.Sprint(page),
		"-l", fmt.Sprint(page)).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 792 // default
	}

	re := regexp.MustCompile("Page\\s+\\d*\\s+size:")
	loc := re.FindIndex(out)
	if loc == nil {
		fmt.Fprintln(os.Stderr, "in pdfHeight: loc == nil")
		return 792
	}

	out = out[loc[1]:]
	i := bytes.Index(out, []byte("x"))
	j := bytes.Index(out, []byte("pts"))
	h, err := strconv.Atoi(string(bytes.Trim(out[i+1:j], " \t")))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 792
	}
	return uint(h)
}

func (ref SheetRef) Get() ([]byte, error) {
    if len(ref) == 0 {
        return nil, errors.New("Invalid reference to image")
    }
    var url = ref[0]
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	if strings.HasSuffix(url, ".pdf") {
		var top []string = nil
		var bottom []string = nil
		if len(ref) > 2 {
			top = []string{ref[1], ref[2]}
		}
		if len(ref) > 4 {
			bottom = []string{ref[3], ref[4]}
		}

		var page = uint64(1)
		if len(ref) > 5 {
			page, _ = strconv.ParseUint(ref[5], 10, 64)
		}

		if _, ok := sheetRefCache[url]; !ok {
			tmp, err := os.Create(".cache/" + strconv.FormatInt(rand.Int63(), 36) + ".pdf")
			if err != nil {
				return nil, err
			}
			tmp.Write(body)
			defer tmp.Close()

			sheetRefCache[url] = tmp.Name()
		}

		var opts = []string{}
		opts = append(opts, "-q", "-dSAFER", "-dBATCH", "-dNOPAUSE", "-sDEVICE=pnggray", "-r144", "-sPageList=" + fmt.Sprint(page), "-sOutputFile=-")

		if len(bottom) >= 2 && len(top) >= 2 {
			topX, _ := strconv.Atoi(top[0])
			topY, _ := strconv.Atoi(top[1])
			bottomX, _ := strconv.Atoi(bottom[0])
			bottomY, _ := strconv.Atoi(bottom[1])

			// TODO
			pageSize := pdfHeight(sheetRefCache[url], uint(page))

			width := bottomX - topX
			height := bottomY - topY

			opts = append(opts, "-dFIXEDMEDIA", "-dDEVICEWIDTHPOINTS=" + fmt.Sprint(width), "-dDEVICEHEIGHTPOINTS=" + fmt.Sprint(height))
			opts = append(opts, "-c",  "<</Install {-" + fmt.Sprint(topX) + " " +  fmt.Sprint(topY - (int(pageSize) - height)) + " translate}>> setpagedevice")
		}

		opts = append(opts, "-f", sheetRefCache[url])
		out, err := exec.Command("gs", opts...).Output()
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintln(os.Stderr, string(exiterr.Stderr))
			}
			return nil, err
		}

		return out, nil
	} else if strings.HasSuffix(url, ".png") {
		return body, nil
	} else {
		return nil, errors.New("Unsupported filetype")
	}
}
