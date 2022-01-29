package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Crawler interface {
	close() error
	getItem(id string) string
	getItemsFromPage(page int) []string
	getNumPages() int
	getPage(id int) (*goquery.Document, error)
	run() (int, []string, error)
}

func crawl(uri string) (*goquery.Document, error) {
	// Request the HTML page.
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	return goquery.NewDocumentFromReader(res.Body)
}

func isMatch(text string) bool {
	if strings.Contains(text, " dogs ") || strings.Contains(text, " pets ") {
		if strings.Contains(text, "no dogs") ||
			strings.Contains(text, "dogs not") ||
			strings.Contains(text, "no pets") ||
			strings.Contains(text, "pets not") {
			return false
		}

		return true
	}

	return false
}
