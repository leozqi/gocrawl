package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
    "flag"
)

type CrawlJob struct {
    url string
}


func getURL(link string) string {
	res, err := http.Get(link)
	defer res.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func parse(data string, parses []string) *Set {
	tkn := html.NewTokenizer(strings.NewReader(data))

	unique := NewSet()

	var toParse bool

	for {
		tt := tkn.Next()
		switch {
		case tt == html.ErrorToken:
			return unique
		case tt == html.StartTagToken:
			toParse = contains(parses, tkn.Token().Data)
			break
		case tt == html.TextToken:
			if toParse {
				words := strings.Fields(tkn.Token().Data)
				for _, word := range words {
					if unique.Has(word) == false {
						unique.Add(word)
					}
				}
			}
			toParse = false
		}
	}
}

func crawl(job *CrawlJob, db *sql.DB) bool {
    tx, err := db.Begin()
    if err != nil {
        log.Fatal(err)
    }

    stmt, err := tx.Prepare(`INSERT INTO webcrawler(link, keywords) values (?, ?)`)
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

	var page string = getURL(job.url)

	words := parse(page, settings)
	for key, _ := range words.list {
		fmt.Println(key)
	}


    _, err = stmt.Exec("www.google.com", "google search")
	if err != nil {
		log.Fatal(err)
	}

    err = tx.Commit()
    if err != nil {
        log.Fatal(err)
    }
    return true
}

func main() {
    // Get command line flags
    urlCmd := flag.String("url", "www.google.com", "URL to start crawling at")
    flag.Parse()

    // Setup database object
    db, err := sql.Open("sqlite3", "./gocrawl.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Prepare jobs
    var job CrawlJob
    job.url = *urlCmd

    // Attempt to initialize the webcrawler table
    if InitDbFile(db) != nil {
        log.Printf("%q: %s\n", err, "could not initialize database table")
        return
    }

    crawl(&job, db);
}

