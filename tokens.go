package gomtch

import (
	"unicode"
)

type Tokens struct {
	values  map[int][]rune
	mapping []int
}

type mapping map[string][]int

func (m mapping) upCreate(v string, index int) mapping {
	if _, ok := m[v]; ok {
		m[v] = append(m[v], index)
	} else {
		m[v] = []int{index}
	}
	return m
}

func makeMapping(tokens []string) mapping {
	m := mapping{}
	var cntr int
	var cntrRef int
	for _, t := range tokens {
		cntrRef = cntr
		startSpecial := getStartSpecial([]rune(t))
		if startSpecial != nil {
			for _, s := range startSpecial {
				m = m.upCreate(string(s), cntr)
				cntr++
			}
			t = t[len(startSpecial):]
			cntrRef = cntr
		}
		endSpecial := getEndSpecial([]rune(t))
		if endSpecial != nil {
			cntrRef = cntr
			for _, s := range endSpecial {
				cntr++
				m = m.upCreate(string(s), cntr)
			}
			t = t[:len(t)-len(endSpecial)]
		}
		m = m.upCreate(t, cntrRef)
		cntr++
	}
	return m
}

func getStartSpecial(token []rune) []rune {
	var special []rune
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return special
		}
		special = append(special, r)
	}
	return nil
}

func getEndSpecial(token []rune) []rune {
	var special []rune
	for i := len(token) - 1; i >= 0; i-- {
		if unicode.IsLetter(token[i]) || unicode.IsNumber(token[i]) {
			return special
		}
		special = append(special, token[i])
	}
	return nil
}

func MakeTokens(v []string) Tokens {
	tokens := Tokens{
		values: map[int][]rune{},
	}
	m := makeMapping(v)
	tokens.mapTokens(m)
	return tokens
}

func (t *Tokens) mapTokens(m mapping) {
	reference := map[string]int{}
	var next int
Outer:
	for {
		for word, indexes := range m {
			for _, index := range indexes {
				if index == next {
					if id, ok := reference[word]; ok {
						t.mapping = append(t.mapping, id)
						next++
						continue Outer
					}
					reference[word] = next
					t.values[next] = []rune(word)
					t.mapping = append(t.mapping, next)
					next++
					continue Outer
				}
			}
		}
		return
	}
}
