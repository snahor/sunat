package sunat

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func fakeResponseFromFile(filename string) *http.Response {
	url := &url.URL{
		Scheme: "http",
		Host:   "www.example.com",
	}
	req := &http.Request{
		URL: url,
	}
	file, _ := os.Open(filename)
	res := &http.Response{
		Request: req,
		Body:    file,
	}
	return res
}

type testHelper struct {
	*testing.T
}

func (t testHelper) assert(title string, expected interface{}, got interface{}) {
	if got != expected {
		t.Errorf("%q: Expected: %v, got: %v", title, expected, got)
	}
}

func Test_hasError(t *testing.T) {
	th := testHelper{t}
	cases := []struct {
		filename string
		hasError bool
	}{
		{"testdata/error.html", true},
		{"testdata/empty_results.html", false},
		{"testdata/results.html", false},
	}
	for _, c := range cases {
		doc, _ := goquery.NewDocumentFromResponse(fakeResponseFromFile(c.filename))
		th.assert("hasError:"+c.filename, hasError(doc) != nil, c.hasError)
	}
}

func Test_parseResults(t *testing.T) {
	th := testHelper{t}
	cases := []struct {
		filename string
		count    int
		total    int
		page     int
	}{
		{"testdata/results.html", 6, 6, 1},
		{"testdata/empty_results.html", 0, 0, 0},
		{"testdata/multipage_results.html", 30, 803, 1},
		{"testdata/multipage_results_p9.html", 30, 803, 9},
		{"testdata/multipage_results_p27.html", 23, 803, 27},
	}
	for _, c := range cases {
		rs, _ := parseResults(fakeResponseFromFile(c.filename))
		th.assert("Filename: "+c.filename, c.count, len(rs.Results))
		th.assert("Filename: "+c.filename, c.total, rs.Metadata.Total)
		th.assert("Filename: "+c.filename, c.page, rs.Metadata.Page)
	}
}

func Test_parseDetail(t *testing.T) {
	th := testHelper{t}
	cases := []struct {
		filename string
		Ruc      string
		Dni      string
		Type     string
		Name     string
	}{
		{
			"testdata/detail.html",
			"20254138577",
			"",
			"SOC.COM.RESPONS. LTDA",
			"MICROSOFT PERU S.R.L.",
		},
		{
			"testdata/person_detail.html",
			"10441233901",
			"44123390",
			"PERSONA NATURAL SIN NEGOCIO",
			"HUMALA TASSO, OLLANTA MOISES",
		},
	}
	for _, c := range cases {
		d, _ := parseDetail(fakeResponseFromFile(c.filename))
		th.assert("parseDetail", c.Ruc, d.Ruc)
		th.assert("parseDetail", c.Dni, d.Dni)
		th.assert("parseDetail", c.Type, d.Type)
		th.assert("parseDetail", c.Name, d.Name)
	}
}

func TestDetailToResult(t *testing.T) {
	th := testHelper{t}
	d, _ := parseDetail(fakeResponseFromFile("testdata/person_detail.html"))
	rs := d.ToResults()
	th.assert("TestDetailToResult", 1, rs.Metadata.Total)
	th.assert("TestDetailToResult", 1, len(rs.Results))
	th.assert("TestDetailToResult", "HUMALA TASSO, OLLANTA MOISES", rs.Results[0].Name)
	th.assert("TestDetailToResult", "10441233901", rs.Results[0].Ruc)
}

func TestGetDetail(t *testing.T) {
	th := testHelper{t}
	d, err := GetDetail("10441233901")
	if err != nil {
		if err != ErrInvalidCaptcha {
			t.Error(err)
		}
	} else {
		th.assert("getDetail()", "HUMALA TASSO, OLLANTA MOISES", d.Name)
		th.assert("getDetail()", "44123390", d.Dni)
	}
	_, err = GetDetail("10441233909")
	th.assert("GetDetail", ErrInvalidRUC, err)
}

func TestSearch(t *testing.T) {
	th := testHelper{t}
	rs, err := Search("OLLANTA HUMALA TASSO")
	if err != nil {
		if err != ErrInvalidCaptcha {
			t.Error(err)
		}
	} else {
		th.assert("Search", 3, rs.Metadata.Total)
		th.assert("Search", 3, len(rs.Results))
	}

	_, err = Search("10441233901")
	th.assert("Search", ErrValueNotSupported, err)
}
