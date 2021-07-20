package main

import (
	"bytes"
	"github.com/nicolasassi/gomtch/document"
	"log"
	"regexp"
)

func main() {
	text := "<p>This is a REAAAAAAL WORLD example of a téxt quite h4rd to match!!<p>"
	corp, err := document.NewDocument(text,
		document.WithSetLower(),
		document.WithSequentialEqualCharsRemoval(),
		document.WithHMTLParsing(),
		document.WithReplacer(regexp.MustCompile(`[\[\]()\-.,:;{}"'!?]`), " "),
		document.WithTransform(document.NewASCII()))
	if err != nil {
		log.Fatal(err)
	}
	match1, err := document.NewDocumentFromReader(bytes.NewReader([]byte("real world")))
	if err != nil {
		log.Fatal(err)
	}
	match2, err := document.NewDocumentFromReader(bytes.NewReader([]byte("text quite hard to match")),
		document.WithMinimumMatchScore(90))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match1, match2) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}
