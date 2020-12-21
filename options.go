package gomtch

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/jdkato/prose/tokenize"
	"regexp"
	"strings"
	"unicode"
)

type Option func(*Doc)

func WithHMTLParsing() Option {
	return func(d *Doc) {
		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(d.Text))
		if err != nil {
			d.optError = err
			return
		}
		d.Text = doc.Text()
	}
}

func WithTransform(t Transformer) Option {
	return func(d *Doc) {
		s, err := t.Transform(d.Text)
		if err != nil {
			d.optError = err
			return
		}
		d.Text = s
	}
}

func WithSequentialEqualCharsRemoval() Option {
	var buf bytes.Buffer
	var pc rune
	return func(d *Doc) {
		for i, c := range d.Text {
			if i == 0 {
				pc = c
				buf.WriteRune(c)
			}
			if pc == c {
				if !unicode.IsNumber(pc) {
					continue
				}
			}
			pc = c
			buf.WriteRune(c)
		}
		d.Text = buf.String()
	}
}

func WithSetLower() Option {
	return func(d *Doc) {
		d.Text = strings.ToLower(d.Text)
	}
}

func WithSetUpper() Option {
	return func(d *Doc) {
		d.Text = strings.ToUpper(d.Text)
	}
}

func WithReplacer(pattern *regexp.Regexp, rep string) Option {
	return func(d *Doc) {
		d.Text = pattern.ReplaceAllString(d.Text, rep)
	}
}

func WithMinimumMatchScore(score int) Option {
	return func(d *Doc) {
		d.matchScoreFunc = func(matchScore, wordLength int) bool {
			return matchScore >= score*wordLength/100
		}
	}
}

func WithConditionalMatchScore(f func(int, int) bool) Option {
	return func(d *Doc) {
		d.matchScoreFunc = f
	}
}

func WithCustomRegexpTokenizer(t *tokenize.RegexpTokenizer) Option {
	return func(d *Doc) {
		if t == nil {
			d.Tokens = []string{regexp.MustCompile(`\s+`).ReplaceAllString(d.Text, "")}
			return
		}
		d.Tokens = t.Tokenize(d.Text)
	}
}
