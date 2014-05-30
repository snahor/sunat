package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

var (
	host = flag.String("host", "127.0.0.1", "host to listen on")
	port = flag.String("port", "8888", "port to listen on")
)

type Results struct {
	Items   []map[string]string `json:"items"`
	Total   int                 `json:"total"`
	PerPage int                 `json:"per_page"`
	Page    int                 `json:"page"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	data, err := Search(q)
	results := Results{data, len(data), 30, 1}
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
	}
	b, _ := json.Marshal(results)
	w.Write(b)
}

func main() {
	if !isTesseractInstalled() {
		log.Fatal("Tesseract is not installed.")
	}
	flag.Parse()
	http.HandleFunc("/", searchHandler)
	http.ListenAndServe(*host+":"+*port, nil)
}
