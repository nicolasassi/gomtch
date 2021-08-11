package gomtch

type Tokens struct {
	Values map[int][]rune
	Ids    []int
}

func (t Tokens) GetRunesByID(id int) []rune {
	return t.Values[id]
}
