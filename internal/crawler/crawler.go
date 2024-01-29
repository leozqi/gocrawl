// Copyright (C) 2024 Leo Qi <leo@leozqi.com>

package crawler

import (
    "database/sql"
    "strings"
    "net/http"
    "golang.org/x/net/html"
    "io/ioutil"
    "bytes"
    "hash/crc32"
    "encoding/hex"

    "gocrawl/internal/graph"
    "gocrawl/internal/utils"
)


type CrawlJob struct {
    Url string
    KeywordTags *utils.Set // HTML tags to search for keywords
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


func ParseDOMString(dom string, keywordTags *utils.Set) (*utils.Set, *utils.Set) {
    tkn := html.NewTokenizer(strings.NewReader(dom))
    unique := utils.NewSet()
    links := utils.NewSet()

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
    pageBytes, err := DownloadPage(job.Url)
    if err != nil {
        return err
    }

    // Get hash of all page contents
    hasher := crc32.NewIEEE()
    hasher.Write(pageBytes)
    hash := hex.EncodeToString(hasher.Sum(nil))

    keywords, links := ParseDOMString(string(pageBytes), job.KeywordTags)

    webpage := graph.Webpage{job.Url, hash, keywords.Slice(), links.Slice()}

    return graph.InsertCrawlResult(db, &webpage)
}
