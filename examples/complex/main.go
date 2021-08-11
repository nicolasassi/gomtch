package main

import (
	"bytes"
	"github.com/nicolasassi/gomtch"
	"log"
	"regexp"
)

func main() {
	text := "<p>This is a REAAAAAAL WORLD example of a téxt quite h4rd to match!!<p>"
	corp, err := gomtch.NewDocument(text,
		gomtch.WithSetLower(),
		gomtch.WithSequentialEqualCharsRemoval(),
		gomtch.WithHMTLParsing(),
		gomtch.WithReplacer(regexp.MustCompile(`[\[\]()\-.,:;{}"'!?]`), " "),
		gomtch.WithTransform(gomtch.NewASCII()))
	if err != nil {
		log.Fatal(err)
	}
	match1, err := gomtch.NewDocumentFromReader(bytes.NewReader([]byte("real world")))
	if err != nil {
		log.Fatal(err)
	}
	match2, err := gomtch.NewDocumentFromReader(bytes.NewReader([]byte("text quite hard to match")),
		gomtch.WithMinimumMatchScore(90))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match1, match2) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}
