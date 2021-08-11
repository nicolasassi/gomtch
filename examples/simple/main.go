package main

import (
	"github.com/nicolasassi/gomtch"
	"log"
)

func main() {
	text := "this is a text c o r p o r a"
	tokenToFind := "corpora"
	corp, err := gomtch.NewDocument(text)
	if err != nil {
		log.Fatal(err)
	}
	match, err := gomtch.NewDocument(tokenToFind)
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}
