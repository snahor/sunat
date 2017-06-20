package sunat

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	searchURL  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/jcrS03Alias"
	captchaURL = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/captcha"
	detailURL  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/DatosRazSocCel.jsp"
	perPage    = 30
)

var (
	timeOut              = time.Duration(10 * time.Second)
	ErrValueNotSupported = errors.New("Value not supported")
	ErrInvalidRUC        = errors.New("Invalid RUC")
	ErrRUCCanNotBeUsed   = errors.New("RUC can not be used")
)

type Result struct {
	Ruc      string `json:"ruc"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
	Href     string `json:"href"`
}

type Metadata struct {
	Total   int `json:"total"`
	PerPage int `json:"per_page"`
	Page    int `json:"page"`
}

type Detail struct {
	Status    string `json:"status"`
	Name      string `json:"name"`
	Ruc       string `json:"ruc"`
	Address   string `json:"address"`
	Type      string `json:"type"`
	Condition string `json:"condition"`
	Dni       string `json:"dni"`
}

type Results struct {
	Results  []Result `json:"data"`
	Metadata Metadata `json:"meta"`
}

func (r *Result) Detail() (*Detail, error) {
	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}
	return getDetail(r.Ruc, client)
}

func Search(q string) (*Results, error) {
	data := url.Values{}
	data.Set("contexto", "ti-it")
	switch {
	case isRuc(q):
		return nil, ErrRUCCanNotBeUsed
	case isDni(q):
		data.Set("accion", "consPorTipdoc")
		data.Set("nrodoc", q)
		data.Set("tipdoc", "1")
	case isName(q):
		data.Set("accion", "consPorRazonSoc")
		data.Set("razSoc", q)
	default:
		return nil, ErrValueNotSupported
	}

	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	number, err := getRandomNumber(client)
	if err != nil {
		return nil, err
	}
	data.Set("numRnd", number)

	// a "better and friendlier" response
	//data.Set("modo", "1")

	res, err := client.PostForm(searchURL, data)
	if err != nil {
		return nil, err
	}

	results, err := parseResults(res)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetDetail(ruc string) (*Detail, error) {
	if !isRuc(ruc) {
		return nil, ErrInvalidRUC
	}

	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	detail, err := getDetail(ruc, client)
	if err != nil {
		return nil, err
	}

	return detail, nil
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeOut)
}

func newHTTPClient() (*http.Client, error) {
	transport := http.Transport{Dial: dialTimeout}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:       jar,
		Transport: &transport,
	}

	res, err := client.Head(searchURL)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return client, nil
}

func getRandomNumber(client *http.Client) (string, error) {
	data := url.Values{}
	data.Set("accion", "random")

	res, err := client.PostForm(captchaURL, data)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func getDetail(ruc string, client *http.Client) (*Detail, error) {
	log.Print("----------")
	log.Print("Requesting details...")
	defer log.Print("----------")
	number, err := getRandomNumber(client)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("numRnd", number)
	data.Set("nroRuc", ruc)
	data.Set("accion", "consPorRuc")
	log.Printf("Params: %v", data)
	res, err := client.PostForm(searchURL, data)
	if err != nil {
		return nil, err
	}

	detail, err := parseDetail(res)
	if err != nil {
		return nil, err
	}

	return detail, nil
}

func hasError(doc *goquery.Document) error {
	result := doc.Find("p.error").First()
	if result.Length() > 0 {
		return errors.New(strings.TrimSpace(result.Text()))
	}
	return nil
}

func parseResults(res *http.Response) (*Results, error) {
	log.Print("Parsing results...")

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	if err = hasError(doc); err != nil {
		return nil, err
	}

	results := &Results{}

	rows := doc.Find("td.beta table tr")

	// length includes the "table header"
	length := rows.Length()

	if length < 2 {
		results.Results = make([]Result, 0)
	} else {
		ns := digitsPattern.FindAllString(doc.Find("td.lnk7").First().Text(), -1)
		if len(ns) == 3 {
			from, _ := strconv.Atoi(ns[0])
			total, _ := strconv.Atoi(ns[2])
			results.Metadata.Page = (from / perPage) + 1
			results.Metadata.Total = total
		}

		results.Metadata.PerPage = perPage

		rows.Slice(1, length).Each(func(i int, s *goquery.Selection) {
			cols := s.Find("td")
			ruc := trim(cols.Eq(0).Find("a").Text())
			results.Results = append(results.Results, Result{
				Ruc:      ruc,
				Name:     trim(cols.Eq(1).Text()),
				Location: trim(cols.Eq(2).Text()),
				Status:   trim(cols.Eq(3).Text()),
				Href:     fmt.Sprintf("http://localhost:8888/detail/%v", ruc),
			})
		})
	}

	return results, nil
}

func parseDetail(res *http.Response) (*Detail, error) {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	if err = hasError(doc); err != nil {
		return nil, err
	}

	detail := &Detail{}
	rows := doc.Find("#print table tr")
	rows.Slice(1, rows.Length()).Each(func(i int, s *goquery.Selection) {
		l := trim(s.Find("td.bgn").Text())
		v := trim(s.Find("td.bg").Text())
		//log.Printf("Label: %q Value: %q", l, v)
		switch {
		case strings.HasPrefix(l, "RUC"):
			xs := strings.Split(v, "-")
			detail.Ruc = trim(xs[0])
			detail.Name = trim(xs[1])
		case strings.HasPrefix(l, "Esta"):
			detail.Status = v
		case strings.HasPrefix(l, "Domi"):
			detail.Address = removeExtraSpaces(v)
		case strings.HasPrefix(l, "Cond"):
			detail.Condition = v
		case strings.HasPrefix(l, "Tipo Con"):
			detail.Type = v
		case strings.HasPrefix(l, "Tipo de Doc"):
			detail.Dni = digitsPattern.FindString(v)
			detail.Name = trim(strings.Split(v, "-")[1])
		}
	})

	return detail, nil
}
