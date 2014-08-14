package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	SearchURL  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/jcrS00Alias"
	CaptchaURL = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/captcha?accion=image"
	DetailURL  = "http://www.sunat.gob.pe/w/wapS01Alias?ruc="
	PagingURL  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/datosRazSoc"
)

var (
	ErrInvalidRUC       = errors.New("Invalid RUC")
	ErrNonexistentRUC   = errors.New("RUC doesn't exist")
	ErrUnexpected       = errors.New("Unexpected error")
	ErrSUNAT            = errors.New("Unexpected error from SUNAT")
	ErrUnsupportedValue = errors.New("Unsupported value")
	ErrCaptcha          = errors.New("Captcha error")
)

var timeout = time.Duration(10 * time.Second)

type Result struct {
	Ruc      string `json:"ruc"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

type Metadata struct {
	Total   int `json:"total"`
	PerPage int `json:"per_page"`
	Page    int `json:"page"`
}

type Results struct {
	Results  []Result `json:"results"`
	Metadata Metadata `json:"_meta"`
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func hasError(doc *goquery.Document) bool {
	result := doc.Find("p.error").First()
	if result.Length() > 0 {
		text := strings.TrimSpace(result.Text())
		log.Print(text)
		return text != ""
	}
	return false
}

func guessCaptcha(client *http.Client) (string, error) {
	resp, err := client.Get(CaptchaURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmpfile, err := ioutil.TempFile("/tmp", "img")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	if _, err := io.Copy(tmpfile, resp.Body); err != nil {
		return "", err
	}

	captcha := captchaToText(tmpfile.Name())
	if err := os.Remove(tmpfile.Name()); err != nil {
		return "", err
	}

	if captcha == "" {
		return "", errors.New("Could not recognize image.")
	}

	return captcha, nil
}

func Search(q string) (*Results, error) {
	postData := url.Values{}

	if isDni(q) {
		postData.Set("accion", "consPorTipdoc")
		postData.Set("nrodoc", q)
		postData.Set("tipdoc", "1")
	} else if isRuc(q) {
		postData.Set("accion", "consPorRuc")
		postData.Set("nroRuc", q)
	} else if isName(q) {
		postData.Set("accion", "consPorRazonSoc")
		postData.Set("razSoc", q)
	} else {
		return nil, ErrUnsupportedValue
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:       cookieJar,
		Transport: &transport,
	}

	resp, err := client.Head(SearchURL)
	if err != nil {
		log.Print(err)
		return nil, ErrSUNAT
	}
	defer resp.Body.Close()

	captcha, err := guessCaptcha(client)
	if err != nil {
		log.Print(err)
		return nil, ErrUnexpected
	}

	postData.Set("codigo", captcha)
	postData.Set("contexto", "ti-it")

	resp2, err := client.PostForm(SearchURL, postData)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp2)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if hasError(doc) {
		return nil, ErrSUNAT
	}

	rows := doc.Find("td.beta table tr")

	results := &Results{}

	// length includes the "table header"
	length := rows.Length()

	if length < 2 {
		results.Results = make([]Result, 0)
	} else {
		re := regexp.MustCompile("(\\d+)$")

		total, err := strconv.Atoi(re.FindString(doc.Find("td.lnk7").First().Text()))
		if err != nil {
			total = 0
		}
		results.Metadata.Total = total
		results.Metadata.PerPage = 30

		rows.Slice(1, length).Each(func(i int, s *goquery.Selection) {
			cols := s.Find("td")
			result := Result{
				// RUC
				strings.TrimSpace(cols.Eq(0).Find("a").Text()),
				// Name
				strings.TrimSpace(cols.Eq(1).Text()),
				// Location
				strings.TrimSpace(cols.Eq(2).Text()),
				// Status
				strings.TrimSpace(cols.Eq(3).Text()),
			}
			results.Results = append(results.Results, result)
		})
	}
	return results, nil
}
