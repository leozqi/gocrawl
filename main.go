package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
    "flag"
)

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
    job.keywordTags = NewSet()
    job.keywordTags.AddMulti("p")

    // Attempt to initialize the webcrawler table
    if InitDbFile(db) != nil {
        log.Printf("%q: %s\n", err, "could not initialize database table")
        return
    }

    Crawl(&job, db);
}

