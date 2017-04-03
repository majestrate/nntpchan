package main

import (
	"bytes"
	"fmt"
	"github.com/whyrusleeping/TinyHtml"
	"io/ioutil"
	"net/http"
)

func GetAndCompress(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	premin := len(body)
	prBuf := bytes.NewBuffer(body)
	tinifier := tinyhtml.NewMinimizer(prBuf)
	min, err := ioutil.ReadAll(tinifier)
	if err != nil {
		fmt.Println(err)
		return
	}
	aftermin := len(min)
	fmt.Printf("Minimized %s\nBy:\t%d bytes\nor\t%f%%\n", url, premin-aftermin, 100.0*float64(premin-aftermin)/float64(premin))
}

func main() {
	fmt.Println("Enter a webpage to test compression on:")
	var url string
	fmt.Scanln(&url)
	GetAndCompress(url)
}
