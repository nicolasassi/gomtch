package main

import (
	"github.com/nicolasassi/gomtch"
	"log"
)

func main() {
	text := "this is a text corp0ra"
	tokenToFind := "corpora"
	corp, err := gomtch.NewDocument(text)
	if err != nil {
		log.Fatal(err)
	}
	match, err := gomtch.NewDocument(tokenToFind, gomtch.WithMinimumMatchScore(90))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}
