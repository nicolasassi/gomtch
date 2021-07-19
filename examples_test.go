package gomtch_test

import (
	"bytes"
	"github.com/nicolasassi/gomtch"
	"log"
	"regexp"
	"testing"
)

func TestSimple(t *testing.T) {
	text := []byte("this is a text c o r p o r a")
	tokenToFind := []byte("corpora")
	corp, err := gomtch.NewDocFromReader(bytes.NewReader(text))
	if err != nil {
		log.Fatal(err)
	}
	match, err := gomtch.NewDocFromReader(bytes.NewReader(tokenToFind))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}

func TestPlayingWithMatchScores(t *testing.T) {
	text := []byte("this is a text corp0ra")
	tokenToFind := []byte("corpora")
	corp, err := gomtch.NewDocFromReader(bytes.NewReader(text))
	if err != nil {
		log.Fatal(err)
	}
	match, err := gomtch.NewDocFromReader(bytes.NewReader(tokenToFind), gomtch.WithMinimumMatchScore(90))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}

func TestComplexExample(t *testing.T) {
	text := []byte("<p>This is REAAAAAAL WORLD example of a téxt quite h4rd to match!!<p>")
	corp, err := gomtch.NewDocFromReader(bytes.NewReader(text),
		gomtch.WithSetLower(),
		gomtch.WithSequentialEqualCharsRemoval(),
		gomtch.WithHMTLParsing(),
		gomtch.WithReplacer(regexp.MustCompile(`[\[\]()\-.,:;{}"'!?]`), " "),
		gomtch.WithTransform(gomtch.NewASCII()))
	if err != nil {
		log.Fatal(err)
	}
	match1, err := gomtch.NewDocFromReader(bytes.NewReader([]byte("real world")))
	if err != nil {
		log.Fatal(err)
	}
	match2, err := gomtch.NewDocFromReader(bytes.NewReader([]byte("text quite hard to match")),
		gomtch.WithMinimumMatchScore(90))
	if err != nil {
		log.Fatal(err)
	}
	for index, match := range corp.Scan(match1, match2) {
		log.Printf("index: %v match: %s", index, string(match))
	}
}
