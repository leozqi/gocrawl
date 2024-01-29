// Copyright (C) 2024 Leo Qi <leo@leozqi.com>

package main

import (
    "database/sql"
    "flag"
    "log"

    _ "github.com/mattn/go-sqlite3"

    "gocrawl/internal/crawler"
    "gocrawl/internal/graph"
    "gocrawl/internal/utils"
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
    var job crawler.CrawlJob
    job.Url = *urlCmd
    job.KeywordTags = utils.NewSet()
    job.KeywordTags.AddMulti("p")

    // Attempt to initialize the webcrawler table
    if graph.InitDbFile(db) != nil {
        log.Printf("%q: %s\n", err, "could not initialize database table")
        return
    }

    err = crawler.Crawl(&job, db)
    if err != nil {
        log.Fatal(err)
    }
}

