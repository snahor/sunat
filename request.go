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
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	search_url  = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/jcrS00Alias"
	captcha_url = "http://www.sunat.gob.pe/cl-ti-itmrconsruc/captcha?accion=image"
	ruc_url     = "http://www.sunat.gob.pe/w/wapS01Alias?ruc="
)

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func getErrorMessage(doc *goquery.Document) string {
	result := doc.Find("p.error").First()
	text := ""
	if result.Length() > 0 {
		text = result.Text()
	}
	return strings.TrimSpace(text)
}

func Search(q string) ([]map[string]string, error) {
	postData := url.Values{}

	if isDni(q) {
		postData.Set("accion", "consPorTipdoc")
		postData.Set("nrodoc", q)
		postData.Set("tipdoc", "1")
	} else if isRuc(q) {
		result, err := GetByRuc(q)
		data := make([]map[string]string, 1)
		data[0] = result
		return data, err
	} else if isName(q) {
		postData.Set("accion", "consPorRazonSoc")
		postData.Set("razSoc", q)
	} else {
		return nil, errors.New("Value not supported.")
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:       cookieJar,
		Transport: &transport,
	}

	resp, err := client.Head(search_url)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()

	resp1, err := client.Get(captcha_url)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer resp1.Body.Close()

	tmpfile, err := ioutil.TempFile("/tmp", "img")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer tmpfile.Close()

	if _, err := io.Copy(tmpfile, resp1.Body); err != nil {
		log.Print(err)
		return nil, err
	}
	captcha := captchaToText(tmpfile.Name())
	if err := os.Remove(tmpfile.Name()); err != nil {
		log.Print(err)
		return nil, err
	}
	if captcha == "" {
		return nil, errors.New("Could not recognize image.")
	}
	postData.Set("codigo", captcha)
	postData.Set("contexto", "ti-it")

	resp2, err := client.PostForm(search_url, postData)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp2)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if errorText := getErrorMessage(doc); errorText != "" {
		return nil, errors.New(errorText)
	}

	rows := doc.Find("td.beta table tr")

	// length includes the "table header"
	length := rows.Length()
	if length < 2 {
		log.Print("There are no results.")
		return nil, nil
	}

	data := make([]map[string]string, length-1)
	rows.Slice(1, length).Each(func(i int, s *goquery.Selection) {
		cols := s.Find("td")
		data[i] = map[string]string{
			"ruc":      strings.TrimSpace(cols.Eq(0).Find("a").Text()),
			"name":     strings.TrimSpace(cols.Eq(1).Text()),
			"location": strings.TrimSpace(cols.Eq(2).Text()),
			"status":   strings.TrimSpace(cols.Eq(3).Text()),
		}
	})
	return data, nil
}

func GetByRuc(ruc string) (map[string]string, error) {
	if !isRuc(ruc) {
		return nil, errors.New("RUC is not valid.")
	}
	resp, _ := http.Get(ruc_url + ruc)
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	if c := doc.Find("#card1 small").First().Length(); c > 0 {
		return nil, errors.New("RUC is not valid.")
	}
	data := map[string]string{
		"ruc": ruc,
	}
	doc.Find("small").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			data["name"] = strings.TrimSpace(strings.Split(strings.SplitN(s.Text(), ".", 2)[1], "-")[1])
		case 3:
			data["status"] = strings.Split(s.Text(), ".")[1]
		case 6:
			data["address"] = strings.TrimSpace(strings.SplitN(s.Text(), ".", 2)[1])
		case 7:
			data["condition"] = strings.TrimSpace(s.Find("b").Text())
		case 10:
			data["type"] = strings.TrimSpace(strings.SplitN(s.Text(), ".", 2)[1])
		case 11:
			data["dni"] = strings.TrimSpace(strings.Split(s.Text(), ":")[1])
		}
	})
	return data, nil
}
