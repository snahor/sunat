package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
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

func searchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.FormValue("q")
	results, err := sunat.Search(query)
	if err == sunat.ErrRUCCanNotBeUsed {
		http.Redirect(w, r, "/detail/"+query, http.StatusSeeOther)
	} else {
		handler(w, results, err)
	}
}

func detailHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	detail, err := sunat.GetDetail(ps.ByName("ruc"))
	handler(w, detail, err)
}

func main() {
	flag.Parse()

	router := httprouter.New()
	router.GET("/search", searchHandler)
	router.GET("/detail/:ruc", detailHandler)

	log.Printf("Listening on: %v", *address)
	log.Fatal(http.ListenAndServe(*address, router))
}
