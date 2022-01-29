package main

import (
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

type Record struct {
	ID      string    `csv:"id"`
	Match   bool      `csv:"match"`
	Visited time.Time `csv:"visited"`
}

type db struct {
	f       *os.File
	records map[string]*Record
}

func newDb(path string) (*db, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
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
