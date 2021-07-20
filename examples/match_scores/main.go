package main

import (
	"github.com/nicolasassi/gomtch/document"
	"log"
)

func main() {
	text := "this is a text corp0ra"
	tokenToFind := "corpora"
	corp, err := document.NewDocument(text)
	if err != nil {
		log.Fatal(err)
	}
	match, err := document.NewDocument(tokenToFind, document.WithMinimumMatchScore(90))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}
