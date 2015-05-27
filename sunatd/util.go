package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"
)

func ServeFormatted(w http.ResponseWriter, r *http.Request, v interface{}) {
	var marshal func(v interface{}) ([]byte, error)

	accept := r.Header.Get("Accept")

	switch accept {
	case "application/xml":
		marshal = xml.Marshal
	default:
		accept = "application/json"
		marshal = json.Marshal
	}

	content, err := marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", accept)
	w.Write(content)
	return
}
