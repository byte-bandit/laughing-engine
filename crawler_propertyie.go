package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type crawlerPropertyIe struct {
	db *db
}

func newPropertyIeCrawler() Crawler {
	db, err := newDb("db2.csv")
	if err != nil {
		log.Fatal(err)
	}

	return &crawlerPropertyIe{
		db: db,
	}
}
func (c *crawlerPropertyIe) close() error {
	return c.db.f.Close()
}

func (c *crawlerPropertyIe) getItem(id string) string {
	doc, err := crawl(fmt.Sprintf("https://www.property.ie/2%s", id))
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	return doc.Find("#searchmoreinfo_description").Text()
}

func (c *crawlerPropertyIe) getItemsFromPage(page int) []string {
	r, _ := regexp.Compile(`\/(\d{1,10})\W*$`)

	doc, err := c.getPage(page)
	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	doc.Find(".search_result").Each(func(i int, s *goquery.Selection) {
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

func (c *crawlerPropertyIe) getNumPages() int {
	doc, err := c.getPage(0)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	pageLinks := doc.Find("#pages").Find("a")
	numPages := pageLinks.Slice(pageLinks.Length()-2, pageLinks.Length()-1) // pick the entry before the "next" link
	nums, err := strconv.Atoi(numPages.Text())
	if err != nil {
		panic(err)
	}

	return nums
}

func (c *crawlerPropertyIe) getPage(id int) (*goquery.Document, error) {
	return crawl(fmt.Sprintf("https://www.property.ie/property-to-let/ireland/sort_date-desc/p_%d/", id))
}

func (c *crawlerPropertyIe) run() (int, []string, error) {
	log.Println("Running PropertyIE crawler ...")
	pages := c.getNumPages()
	log.Printf("Found listings worth %d pages\n", pages)

	var matches []string
	cnt := 0
	for page := 1; page <= pages; page++ {
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
				match := fmt.Sprintf("https://property.ie/2%s", v)
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
