// Copyright (C) 2024 Leo Qi <leo@leozqi.com>
//
// Use of this source code is governed by the Apache-2.0 License.
// Full text can be found in the LICENSE file

package main

import (
    "database/sql"
    "strings"
    "net/http"
    "golang.org/x/net/html"
    "io/ioutil"
    "bytes"
    "hash/crc32"
    "encoding/hex"
)

type CrawlJob struct {
    url string
    keywordTags *Set // HTML tags to search for keywords
}


func DownloadPage(url string) ([]byte, error) {
    // Get Page
    res, err := http.Get(url)
    if err != nil {
        return []byte{}, err
    }
    defer res.Body.Close()

    content, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return []byte{}, err
    }

    return content, nil
}

// Naive way to trim punctuation from keywords
const TRIMMED string = "~`!@#$%^&*()[]{}-_+=\\|:;\"'<>,.?/ \n\t"
const MIN_LEN int = 2


func ParseDOMString(dom string, keywordTags *Set) (*Set, *Set) {
    tkn := html.NewTokenizer(strings.NewReader(dom))
    unique := NewSet()
    links := NewSet()

    var parseElement bool

    for {
        tt := tkn.Next()
        switch {
        case tt == html.ErrorToken:
            return unique, links
        case tt == html.StartTagToken:
            tn, _ := tkn.TagName()

            parseElement = keywordTags.Has(tkn.Token().Data) // tag name

            if len(tn) == 1 && tn[0] == 'a' {
                for {
                    key, val, moreAttr := tkn.TagAttr()
                    link := string(val)

                    if bytes.Equal(key, []byte("href")) && !links.Has(link) {
                        links.Add(link)
                    }

                    if !moreAttr {
                        break
                    }
                }
            }
            break
        case tt == html.TextToken:
            if parseElement {
                splits := strings.Fields(tkn.Token().Data)
                for _, splitChars := range(splits) {
                    word := strings.Trim(splitChars, TRIMMED)
                    if len([]rune(word)) > MIN_LEN && !unique.Has(word) {
                        unique.Add(word)
                    }
                }
            }
            parseElement = false
            break
        }
    }

    return unique, links
}


func Crawl(job *CrawlJob, db *sql.DB) error {
    pageBytes, err := DownloadPage(job.url)
    if err != nil {
        return err
    }

    // Get hash of all page contents
    hasher := crc32.NewIEEE()
    hasher.Write(pageBytes)
    hash := hex.EncodeToString(hasher.Sum(nil))

    keywords, links := ParseDOMString(string(pageBytes), job.keywordTags)

    var w Webpage
    w.id = job.url
    w.checksum = hash
    w.keywords = keywords.Slice()
    w.links = links.Slice()

    return InsertCrawlResult(db, &w)
}
