package main

import (
	"flag"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	host = flag.String("host", "127.0.0.1", "host to listen on")
	port = flag.String("port", "8888", "port to listen on")
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	results, _ := Search(r.FormValue("q"))
	ServeFormatted(w, r, results)
}

func DetailHandler(w http.ResponseWriter, r *http.Request) {
	detail, err := RUCDetail(mux.Vars(r)["ruc"])
	if err != nil {
		ServeFormatted(w, r, err)
	}
	ServeFormatted(w, r, detail)
}

func main() {
	if !isTesseractInstalled() {
		panic("Tesseract is not installed.")
	}

	r := mux.NewRouter()
	r.HandleFunc("/search", SearchHandler)
	r.HandleFunc("/detail/{ruc:\\d{11}}", DetailHandler)
	http.Handle("/", r)

	flag.Parse()

	if err := http.ListenAndServe(*host+":"+*port, nil); err != nil {
		println("Couldn't start the server.")
		println("Reason:", err.Error())
	}
}
