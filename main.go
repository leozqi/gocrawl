package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "golang.org/x/net/html"
    "encoding/json"
    "strings"
)

func getHTML(link string) (string) {
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

func parseHTML(data string, parses []string) []string {
    tkn := html.NewTokenizer(strings.NewReader(data))
    var vals []string
    var toParse bool

    for {
        tt := tkn.Next()
        switch {
        case tt == html.ErrorToken:
            return vals
        case tt == html.StartTagToken:
            toParse = contains(parses, tkn.Token().Data)
            break
        case tt == html.TextToken:
            if toParse {
                vals = append(vals, tkn.Token().Data)
            }
            toParse = false
        }
    }
}

func main() {
    loadedJSON := []byte(`
        [
            "p",
            "b",
            "i",
            "code",
            "pre",
            "span"
        ]
    `)

    var settings []string
    fmt.Println(loadedJSON)

    if err := json.Unmarshal(loadedJSON, &settings); err != nil {
        log.Fatal(err)
    }

    var page string = getHTML("https://stackoverflow.com/questions/38867692/how-to-parse-json-array-in-go")
    fmt.Println(settings)
    fmt.Println(parseHTML(page, settings))
}

