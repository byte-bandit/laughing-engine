package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
)

type Record struct {
	ID      string    `csv:"id"`
	Match   bool      `csv:"match"`
	Visited time.Time `csv:"visited"`
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

func getItem(id string) string {
	doc, err := crawl(fmt.Sprintf("https://www.propertypal.com/%s", id))
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	return doc.Find(".prop-descr-text").Text()
}

func getPage(id int) (*goquery.Document, error) {
	return crawl(fmt.Sprintf("https://www.propertypal.com/search?sta=toLet&st=rent&max=1500&currency=EUR&minbeds=2&sort=dateHigh&excludePoa=true&pt=residential&stygrp=10&stygrp=8&stygrp=44&excatt=20&page=%d", id))
}

func getItemsFromPage(page int) []string {
	r, _ := regexp.Compile(`\/(\d{1,10})\W*$`)

	doc, err := getPage(page)
	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	doc.Find(".propbox").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		lnk, _ := s.Find("a").Last().Attr("href")
		matches := (r.FindStringSubmatch(lnk))
		if len(matches) != 2 {
			log.Printf("!! Skipping invalid link: %s\n", lnk)
			return
		}

		ids = append(ids, strings.TrimSpace(matches[1]))
	})

	return ids
}

func getNumPages() int {
	doc, err := getPage(0)
	if err != nil {
		log.Fatal(err)
	}

	return doc.Find(".paging-page").Length()

}

func main() {
	db, err := newDb()
	if err != nil {
		log.Fatal(err)
	}
	defer db.close()

	pages := getNumPages()
	log.Printf("Found listings worth %d pages\n", pages)

	var matches []string
	cnt := 0
	for page := 0; page < pages; page++ {
		log.Printf("Reading page %d\n", page)
		items := getItemsFromPage(page)
		log.Printf("Found %d items\n", len(items))
		for _, v := range items {
			if db.get(v) != nil {
				log.Printf("Item %v already visited, skipping ...", v)
				continue
			}
			log.Printf("Checking item %s\n", v)
			cnt++
			item := getItem(v)
			isMatch := false
			scan := strings.ToLower(item)
			if strings.Contains(scan, "dogs") || strings.Contains(scan, "pets") {
				match := fmt.Sprintf("https://propertypal.com/%s", v)
				matches = append(matches, match)
				log.Printf("MATCH FOUND, check %s \n", match)
				isMatch = true
			}
			db.create(v, isMatch)
		}
	}
	log.Println()
	log.Printf("Mission complete. Checked %d entries.", cnt)
	log.Println("Matches:")
	for _, v := range matches {
		log.Println(v)
	}
	log.Println()
	log.Println("Saving to db ...")
	if err = db.commit(); err != nil {
		log.Fatal(err)
	}
	log.Println("Done!")

}

type db struct {
	f       *os.File
	records map[string]*Record
}

func newDb() (*db, error) {
	file, err := os.OpenFile("db.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	lookup := make(map[string]*Record)
	if info.Size() > 0 {
		var records []*Record
		if err := gocsv.UnmarshalFile(file, &records); err != nil {
			return nil, err
		}

		for _, v := range records {
			lookup[v.ID] = v
		}
	}

	return &db{
		f:       file,
		records: lookup,
	}, nil
}

func (d *db) close() {
	d.f.Close()
}

func (d *db) get(id string) *Record {
	record, ok := d.records[id]
	if !ok {
		return nil
	}
	return record
}

func (d *db) create(id string, match bool) {
	d.records[id] = &Record{
		ID:      id,
		Match:   match,
		Visited: time.Now(),
	}
}

func (d *db) commit() error {
	if _, err := d.f.Seek(0, 0); err != nil {
		return err
	}

	var records []*Record
	for _, v := range d.records {
		records = append(records, v)
	}

	return gocsv.MarshalFile(&records, d.f)
}
