package sunat

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
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	searchURL  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/jcrS03Alias"
	captchaURL = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/captcha?accion=image"
	detailURL  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/DatosRazSocCel.jsp"
	perPage    = 30
)

var (
	timeOut              = time.Duration(10 * time.Second)
	ErrInvalidCaptcha    = errors.New("Invalid Captcha")
	ErrValueNotSupported = errors.New("Value not supported")
	ErrInvalidRUC        = errors.New("Invalid RUC")
)

type Result struct {
	Ruc      string `json:"ruc"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
	//c        *http.Client
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

func (d *Detail) ToResults() *Results {
	rs := &Results{}
	rs.Metadata.Total = 1
	rs.Metadata.PerPage = perPage
	rs.Results = []Result{
		Result{
			Name:     d.Name,
			Ruc:      d.Ruc,
			Location: d.Address,
			Status:   d.Status,
		},
	}
	return rs
}

func (r *Result) Detail() (*Detail, error) {
	client, _ := newHttpClient()
	return getDetail(r.Ruc, client)
}

func Search(q string) (*Results, error) {
	data := make(url.Values)
	data.Set("contexto", "ti-it")
	switch {
	case isDni(q):
		data.Set("accion", "consPorTipdoc")
		data.Set("nrodoc", q)
		data.Set("tipdoc", "1")
	case isName(q):
		data.Set("accion", "consPorRazonSoc")
		data.Set("razSoc", q)
	case true:
		return nil, ErrValueNotSupported
	}

	client, err := newHttpClient()
	if err != nil {
		return nil, err
	}

	captcha, err := getCaptcha(client)
	if err != nil {
		return nil, err
	}
	data.Set("codigo", captcha)

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

func GetDetail(q string) (*Detail, error) {
	if !isRuc(q) {
		return nil, ErrInvalidRUC
	}

	client, err := newHttpClient()
	if err != nil {
		return nil, err
	}

	detail, err := getDetail(q, client)
	if err != nil {
		return nil, err
	}

	return detail, nil
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeOut)
}

func newHttpClient() (*http.Client, error) {
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

func getCaptcha(c *http.Client) (string, error) {
	log.Print("Getting captcha...")
	res, err := c.Get(captchaURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	file, err := ioutil.TempFile("/tmp", "img")
	if err != nil {
		return "", err
	}

	// write image
	if _, err := io.Copy(file, res.Body); err != nil {
		return "", err
	}
	log.Printf("Closing temp image: %q...", file.Name())
	defer file.Close()

	text, err := captchaToText(file.Name())

	if err != nil {
		return "", err
	}

	log.Printf("Captcha: %q", text)

	if !isValidCaptcha(text) {
		return "", ErrInvalidCaptcha
	}

	// remove the image only if everythng went well
	log.Printf("Removing temp image: %q...", file.Name())
	defer os.Remove(file.Name())

	return text, nil
}

func getDetail(ruc string, client *http.Client) (*Detail, error) {
	data := make(url.Values)
	data.Set("contexto", "ti-it")
	data.Set("accion", "consPorRuc")
	data.Set("nroRuc", ruc)

	captcha, err := getCaptcha(client)
	if err != nil {
		return nil, ErrInvalidCaptcha
	}
	data.Set("codigo", captcha)

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

func captchaToText(path string) (string, error) {
	output, err := exec.Command(
		"tesseract",
		path,
		"stdout",
		"-psm", "8",
		"-c", "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	).Output()

	if err != nil {
		return "", err
	}

	return string(output[:4]), nil
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
			results.Results = append(results.Results, Result{
				Ruc:      trim(cols.Eq(0).Find("a").Text()),
				Name:     trim(cols.Eq(1).Text()),
				Location: trim(cols.Eq(2).Text()),
				Status:   trim(cols.Eq(3).Text()),
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
