package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Detail struct {
	Status    string `json:"status"`
	Name      string `json:"name"`
	Ruc       string `json:"ruc"`
	Address   string `json:"address"`
	Type      string `json:"type"`
	Condition string `json:"condition"`
	Dni       string `json:"dni"`
}

func RUCDetail(ruc string) (*Detail, error) {
	if !isRuc(ruc) {
		return nil, ErrInvalidRUC
	}

	response, _ := http.Get(DetailURL + ruc)
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		log.Print(err)
		return nil, ErrUnexpected
	}

	// If present the RUC doesn't exist
	if c := doc.Find("#card1 small").First().Length(); c > 0 {
		return nil, ErrNonexistentRUC
	}

	detail := &Detail{Ruc: ruc}

	doc.Find("small").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			detail.Name = strings.TrimSpace(strings.Split(strings.SplitN(s.Text(), ".", 2)[1], "-")[1])
		case 3:
			detail.Status = strings.Split(s.Text(), ".")[1]
		case 6:
			detail.Address = strings.TrimSpace(strings.SplitN(s.Text(), ".", 2)[1])
		case 7:
			detail.Condition = strings.TrimSpace(s.Find("b").Text())
		case 10:
			detail.Type = strings.TrimSpace(strings.SplitN(s.Text(), ".", 2)[1])
		case 11:
			detail.Dni = strings.TrimSpace(strings.Split(s.Text(), ":")[1])
		}
	})
	return detail, nil
}
