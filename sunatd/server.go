package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/snahor/sunat"
)

var (
	address = flag.String("address", ":8888", "HTTP Listener address")
)

func handler(w http.ResponseWriter, v interface{}, err error) {
	var content []byte
	code := http.StatusOK

	if err != nil {
		switch err {
		case sunat.ErrInvalidRUC, sunat.ErrValueNotSupported:
			code = http.StatusBadRequest
		default:
			code = http.StatusInternalServerError
		}

		content, _ = json.Marshal(map[string]string{"error": err.Error()})
	} else {
		content, _ = json.Marshal(v)
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(content)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	results, err := sunat.Search(r.FormValue("q"))
	handler(w, results, err)
}

func DetailHandler(w http.ResponseWriter, r *http.Request) {
	detail, err := sunat.GetDetail(mux.Vars(r)["ruc"])
	handler(w, detail, err)
}

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/search", SearchHandler)
	r.HandleFunc("/detail/{ruc:\\d{11}}", DetailHandler)
	http.Handle("/", r)

	log.Printf("Listening on: %v", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}
