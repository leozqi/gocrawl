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
)

type Set struct {
	list map[string]struct{} //empty structs occupy 0 memory
}

func (s *Set) Has(v string) bool {
	_, ok := s.list[v]
	return ok
}

func (s *Set) Add(v string) {
	s.list[v] = struct{}{}
}

func (s *Set) Remove(v string) {
	delete(s.list, v)
}

func (s *Set) Clear() {
	s.list = make(map[string]struct{})
}

func (s *Set) Size() int {
	return len(s.list)
}

func NewSet() *Set {
	s := &Set{}
	s.list = make(map[string]struct{})
	return s
}

//optional functionalities

// AddMulti Add multiple values in the set
func (s *Set) AddMulti(list ...string) {
	for _, v := range list {
		s.Add(v)
	}
}

func (s *Set) Union(s2 *Set) *Set {
	res := NewSet()
	for v := range s.list {
		res.Add(v)
	}

	for v := range s2.list {
		res.Add(v)
	}
	return res
}

func (s *Set) Intersect(s2 *Set) *Set {
	res := NewSet()
	for v := range s.list {
		if s2.Has(v) == false {
			continue
		}
		res.Add(v)
	}
	return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param)
func (s *Set) Difference(s2 *Set) *Set {
	res := NewSet()
	for v := range s.list {
		if s2.Has(v) {
			continue
		}
		res.Add(v)
	}
	return res
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

func crawl(settings []string, db *sql.DB) bool {
    tx, err := db.Begin()
    if err != nil {
        log.Fatal(err)
    }

    stmt, err := tx.Prepare(`INSERT INTO webcrawler(link, keywords) values (?, ?)`)
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

	var page string = getURL("https://stackoverflow.com/questions/38867692/how-to-parse-json-array-in-go")

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

	// load DB
	db, err := sql.Open("sqlite3", "./search.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
        CREATE VIRTUAL TABLE IF NOT EXISTS webcrawler USING fts5 (
            link,
            keywords
        );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

    var settings []string
	if err := json.Unmarshal(loadedJSON, &settings); err != nil {
		log.Fatal(err)
	}

    crawl(settings[:], db);
}

