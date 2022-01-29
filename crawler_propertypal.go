package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type crawlerPropertyPal struct {
	db *db
}

func newPropertyPalCrawler() Crawler {
	db, err := newDb("db.csv")
	if err != nil {
		log.Fatal(err)
	}

	return &crawlerPropertyPal{
		db: db,
	}
}
func (c *crawlerPropertyPal) close() error {
	return c.db.f.Close()
}

func (c *crawlerPropertyPal) getItem(id string) string {
	doc, err := crawl(fmt.Sprintf("https://www.propertypal.com/%s", id))
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	return doc.Find(".prop-descr-text").Text()
}

func (c *crawlerPropertyPal) getItemsFromPage(page int) []string {
	r, _ := regexp.Compile(`\/(\d{1,10})\W*$`)

	doc, err := c.getPage(page)
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

func (c *crawlerPropertyPal) getNumPages() int {
	doc, err := c.getPage(0)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	return doc.Find(".paging-page").Length()
}

func (c *crawlerPropertyPal) getPage(id int) (*goquery.Document, error) {
	return crawl(fmt.Sprintf("https://www.propertypal.com/search?sta=toLet&st=rent&max=1500&currency=EUR&minbeds=2&sort=dateHigh&excludePoa=true&pt=residential&stygrp=10&stygrp=8&stygrp=44&excatt=20&page=%d", id))
}

func (c *crawlerPropertyPal) run() (int, []string, error) {
	log.Println("Running PropertyPal crawler ...")
	pages := c.getNumPages()
	log.Printf("Found listings worth %d pages\n", pages)

	var matches []string
	cnt := 0
	for page := 0; page < pages; page++ {
		log.Printf("Reading page %d\n", page)
		items := c.getItemsFromPage(page)
		log.Printf("Found %d items\n", len(items))
		for _, v := range items {
			if c.db.get(v) != nil {
				log.Printf("Item %v already visited, skipping ...", v)
				continue
			}
			log.Printf("Checking item %s\n", v)
			cnt++
			item := c.getItem(v)
			isMatched := false
			scan := strings.ToLower(item)
			if isMatch(scan) {
				match := fmt.Sprintf("https://propertypal.com/%s", v)
				matches = append(matches, match)
				log.Printf("MATCH FOUND, check %s \n", match)
				isMatched = true
			}
			c.db.create(v, isMatched)
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
	if err := c.db.commit(); err != nil {
		return 0, nil, err
	}

	log.Println("Done!")
	return cnt, matches, nil
}
