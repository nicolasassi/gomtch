package document

import (
	"fmt"
	"github.com/nicolasassi/gomtch/mapper"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
)

const whiteSpace = ' '

var (
	numericalInfo = []rune("%xª°º")
)

type Documenter interface {
	Compare(ref mapper.Tokens) (bool, []rune)
	IsEqual(a, b []rune) bool
	CompareRune(a, b rune) bool
	fmt.Stringer
}

type Scanner interface {
	Scan(docs ...Documenter) Matches
}

type Document struct {
	matchScoreFunc func(int, int) bool
	transformer    transform.Transformer
	optError       error
	Text           string
	Tokens         []string
}

type Matches map[int][]rune

func NewDocument(text string, opts ...Option) (*Document, error) {
	d := &Document{
		Text: text,
	}
	d.applyOptions(opts...)
	if d.optError != nil {
		return nil, d.optError
	}
	return d, nil
}

func NewDocumentFromReader(text io.Reader, opts ...Option) (*Document, error) {
	b, err := ioutil.ReadAll(text)
	if err != nil {
		return nil, err
	}
	d := &Document{
		Text: string(b),
	}
	d.applyOptions(opts...)
	if d.optError != nil {
		return nil, d.optError
	}
	return d, nil
}

func (d *Document) applyOptions(opts ...Option) {
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *Document as the argument
		opt(d)
		if d.optError != nil {
			return
		}
	}
	if d.Tokens == nil {
		d.Tokens = strings.Split(d.Text, " ")
	}
	if d.matchScoreFunc == nil {
		WithMinimumMatchScore(100)(d)
	}
}

func (d Document) String() string {
	return d.Text
}

// CompareRune compares one rune to another and returns true if there is a match.
// Besides checking equality by the standard form (==) it also applies some rules
// to check if the compared value might be the same as the reference but is masked somehow.
// Numbers and numerical info must match exactly. Also if both A and B are letters both
// should match as well. If the compared entities is not a letter, number or numerical info and
// the reference is not a number or numerical info, it will match.
func (d Document) CompareRune(a, b rune) bool {
	if unicode.IsNumber(a) || isNumericalInfo(a) {
		return a == b
	}
	if unicode.IsLetter(a) && unicode.IsLetter(b) {
		return a == b
	}
	if unicode.IsNumber(b) || isNumericalInfo(b) {
		return false
	}
	if unicode.IsLetter(a) && !unicode.IsLetter(b) {
		return false
	}
	return true
}

// IsEqual compares A and B and returns true if they probably are the same word and false otherwise.
// if A and B have different lengths it will return false.
// The A and B variables are not interchangeable as A represents the entities to be compared to B, the reference.
// IsEqual uses the minimumMatchScore to determine if the words are the same even if there are differences
// between then.
// Numbers, numericalInfo and letters should match exactly and the matches increase the counter for
// the minimumMatchScore, otherwise it returns false immediately.
func (d Document) IsEqual(a, b []rune) bool {
	if len(a) == 1 && len(b) == 1 {
		return d.CompareRune(a[0], b[0])
	}
	if len(a) != len(b) {
		return false
	}
	var matchScore int
	for i, v := range b {
		if v != a[i] {
			// A word made solely of numbers or number related points (%ª°x) can pass only
			// if all points match.
			if !unicode.IsNumber(v) && !isNumericalInfo(v) {
				// if the reference point is not a number or numerical info and its type (letter)
				// does not match the type expected for that point index than the words are not the same.
				if unicode.IsLetter(v) == unicode.IsLetter(a[i]) {
					return false
				}
				continue
			}
			// comparison rules could be added here
			// for now lets keep it simple
			return false
		}
		matchScore++
	}
	return d.matchScoreFunc(matchScore, len(b))
}

func (d Document) Scan(docs ...Documenter) Matches {
	matches := map[int][]rune{}
	tokens := mapper.NewMappingFromTokens(d.Tokens).Map()
	for i, doc := range docs {
		if ok, sequence := doc.Compare(tokens); ok {
			matches[i] = sequence
		}
	}
	return matches
}

type compareEntity struct {
	found    bool
	sequence []rune
	at       int
	last     int
}

func (d Document) Compare(tokens mapper.Tokens) (bool, []rune) {
	ce := compareEntity{}
	var sequence []rune
	for _, ref := range d.Tokens {
		ce.found, sequence, ce.at = d.simpleCheck([]rune(ref), tokens, ce.at)
		if !ce.found {
			break
		}
		if ce.at != ce.last {
			ce.last = ce.at
			ce.at++
		}
		if ce.at == 0 {
			ce.at++
		}

		if ce.sequence == nil {
			ce.sequence = sequence
		} else {
			ce.sequence = append(ce.sequence, whiteSpace)
			ce.sequence = append(ce.sequence, sequence...)
		}
	}
	if !ce.found {
		ce.found, sequence = d.specialCheck([]rune(strings.Join(d.Tokens, "")), tokens)
		if !ce.found {
			return false, nil
		}
		if ce.sequence == nil {
			ce.sequence = sequence
		} else {
			ce.sequence = append(ce.sequence, whiteSpace)
			ce.sequence = append(ce.sequence, sequence...)
		}
	}
	return true, ce.sequence
}

func isSpecial(r rune) bool {
	if unicode.IsNumber(r) || unicode.IsLetter(r) || isNumericalInfo(r) {
		return false
	}
	return true
}

func (d Document) simpleCheck(value []rune, tokens mapper.Tokens, start int) (bool, []rune, int) {
	for i, id := range tokens.Ids {
		if start == 0 {
			if d.IsEqual(tokens.GetRunesByID(id), value) {
				return true, tokens.GetRunesByID(id), i
			}
			continue
		}
		if i < start {
			continue
		}
		if d.IsEqual(tokens.GetRunesByID(id), value) {
			return true, tokens.GetRunesByID(id), i
		} else {
			return false, nil, i
		}
	}
	return false, nil, 0
}

type specialCheck struct {
	startAt            int
	ref                []rune
	compare            []rune
	matchOnOneSpecial  int
	lastFound          bool
	cntr               int
	completeWordSpaced []rune
	completeWord       []rune
}

func newSpecialCheck() *specialCheck {
	return &specialCheck{
		cntr: 1,
	}
}

func (sc *specialCheck) appendCompleteWord(new []rune) {
	if len(sc.completeWordSpaced) == 0 {
		sc.completeWordSpaced = new
		sc.completeWord = new
		return
	}
	sc.completeWordSpaced = append(sc.completeWordSpaced, whiteSpace)
	sc.completeWordSpaced = append(sc.completeWordSpaced, new...)
	sc.completeWord = append(sc.completeWord, new...)
}

func (d Document) specialCheck(value []rune, tokens mapper.Tokens) (bool, []rune) {
	sc := newSpecialCheck()
	for _, id := range tokens.Ids {
	Outer:
		for {
			for i := 0; i < len(value); i++ {
				if i < sc.startAt {
					continue
				}
				sc.ref = value[i : i+sc.cntr]
				sc.compare = tokens.GetRunesByID(id)
				if len(sc.ref) != len(sc.compare) {
					if i+sc.cntr == len(value) {
						sc.lastFound = false
						sc.startAt = 0
						sc.cntr = 1
						break Outer
					}
					sc.cntr++
					break
				}
				if !d.IsEqual(sc.compare, sc.ref) {
					sc.lastFound = false
					sc.startAt = 0
					sc.cntr = 1
					if sc.matchOnOneSpecial != 0 {
						sc.completeWordSpaced = sc.completeWordSpaced[sc.matchOnOneSpecial:]
						sc.completeWord = sc.completeWord[sc.matchOnOneSpecial:]
						sc.matchOnOneSpecial = 0
						continue Outer
					}
					sc.completeWordSpaced = nil
					sc.completeWord = nil
					break Outer
				}
				sc.appendCompleteWord(sc.compare)
				sc.lastFound = true
				if len(sc.ref) == 1 && sc.startAt == 0 {
					if isSpecial(sc.compare[0]) {
						sc.matchOnOneSpecial++
					}
				} else {
					if sc.matchOnOneSpecial != 0 {
						if isSpecial(sc.compare[0]) {
							sc.matchOnOneSpecial++
						} else {
							sc.matchOnOneSpecial = 0
						}
					}
				}
				sc.startAt = i + sc.cntr
				if sc.startAt == len(value) {
					if !d.IsEqual(sc.completeWord, value) {
						return false, nil
					}
					return true, sc.completeWordSpaced
				}
				sc.cntr = 1
				break Outer
			}
		}
	}
	return false, nil
}

func isNumericalInfo(v rune) bool {
	for _, r := range numericalInfo {
		if v == r {
			return true
		}
	}
	return false
}
