package gomtch

import (
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
)

var (
	numericalInfo = []rune("%xª°º")
	whiteSpace    = ' '
)

type Doc struct {
	matchScoreFunc func(int, int) bool
	t              transform.Transformer
	optError       error
	Text           string
	Tokens         []string
}

type Matches map[int][]rune

func NewDoc(text io.Reader, opts ...Option) (*Doc, error) {
	b, err := ioutil.ReadAll(text)
	if err != nil {
		return nil, err
	}
	d := &Doc{
		Text: string(b),
	}
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *Doc as the argument
		opt(d)
	}
	if d.optError != nil {
		return nil, d.optError
	}
	if d.Tokens == nil {
		d.Tokens = strings.Split(d.Text, " ")
	}
	if d.matchScoreFunc == nil {
		WithMinimumMatchScore(100)(d)
	}
	return d, nil
}

func (d Doc) String() string {
	return d.Text
}

// CompareRune compares one rune to another and returns true if they match.
// Besides checking equality by the standard form == it also applies some rules
// to check if the compared value might be the same as the reference but is masked somehow.
// Numbers and numerical info must match exactly. Also if both A and B are letters both
// should match as well. If the compared entities is not a letter, number or numerical info and
// the reference is not a number or numerical info, it will match.
func (d Doc) CompareRune(a, b rune) bool {
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

// IsSame compares A and B and returns true if they probably are the same word and false otherwise.
// if A and B have different lengths it will return false.
// The A and B variables are not interchangeable as A represents the entities to be compared to B, the reference.
// IsSame uses the minimumMatchScore to determine if the words are the same even if there are differences
// between then.
// Numbers, numericalInfo and letters should match exactly and the matches increase the counter for
// the minimumMatchScore, otherwise it returns false immediately.
func (d Doc) IsSame(a, b []rune) bool {
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

func (d Doc) Scan(docs ...Document) Matches {
	m := map[int][]rune{}
	tokens := MakeTokens(d.Tokens)
	for i, doc := range docs {
		if ok, sequence := doc.Compare(tokens); ok {
			m[i] = sequence
		}
	}
	return m
}

type compareEntity struct {
	found    bool
	sequence []rune
	at       int
	last     int
}

func (d Doc) Compare(tokens Tokens) (bool, []rune) {
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

func (d Doc) simpleCheck(value []rune, tokens Tokens, start int) (bool, []rune, int) {
	for i, id := range tokens.mapping {
		if start == 0 {
			if d.IsSame(tokens.values[id], value) {
				//log.Printf("[SIMPLE] compare: %s == ref: %s", string(tokens.values[id]), string(value))
				return true, tokens.values[id], i
			}
			continue
		}
		if i < start {
			continue
		}
		if d.IsSame(tokens.values[id], value) {
			return true, tokens.values[id], i
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

func (d Doc) specialCheck(value []rune, tokens Tokens) (bool, []rune) {
	sc := newSpecialCheck()
	for _, id := range tokens.mapping {
		//log.Printf("word: %s", string(tokens.values[id]))
	Outer:
		for {
			for i := 0; i < len(value); i++ {
				if i < sc.startAt {
					continue
				}
				sc.ref = value[i : i+sc.cntr]
				sc.compare = tokens.values[id]
				//log.Printf("sc.ref: %v sc.compare: %v", string(sc.ref), string(tokens.values[id]))
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
				//log.Printf("[SPECIAL] compare: %s and ref: %s", string(sc.compare), string(sc.ref))
				if !d.IsSame(sc.compare, sc.ref) {
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
				//log.Printf("[SPECIAL] compare: %s == ref: %s", string(sc.compare), string(sc.ref))
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
					if !d.IsSame(sc.completeWord, value) {
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
