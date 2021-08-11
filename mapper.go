package gomtch

import "unicode"

type Mapper interface {
	Map() Tokens
	Items() map[string][]int
	AddIndex(v string, index int) Mapping
}

type Mapping map[string][]int

func NewMappingFromTokens(tokens []string) Mapping {
	m := Mapping{}
	var cntr int
	var cntrRef int
	for _, t := range tokens {
		cntrRef = cntr
		startSpecial := getStartSpecial([]rune(t))
		if startSpecial != nil {
			for _, s := range startSpecial {
				m = m.AddIndex(string(s), cntr)
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
				m = m.AddIndex(string(s), cntr)
			}
			t = t[:len(t)-len(endSpecial)]
		}
		m = m.AddIndex(t, cntrRef)
		cntr++
	}
	return m
}

func (m Mapping) Map() Tokens {
	tokens := Tokens{
		Values: map[int][]rune{},
	}
	reference := map[string]int{}
	var next int
Outer:
	for {
		for word, indexes := range m.Items() {
			for _, index := range indexes {
				if index == next {
					if id, ok := reference[word]; ok {
						tokens.Ids = append(tokens.Ids, id)
						next++
						continue Outer
					}
					reference[word] = next
					tokens.Values[next] = []rune(word)
					tokens.Ids = append(tokens.Ids, next)
					next++
					continue Outer
				}
			}
		}
		break
	}
	return tokens
}

func (m Mapping) Items() map[string][]int {
	return m
}

func (m Mapping) AddIndex(v string, index int) Mapping {
	if _, ok := m[v]; ok {
		m[v] = append(m[v], index)
	} else {
		m[v] = []int{index}
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
