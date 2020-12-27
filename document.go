package gomtch

import "fmt"

type Document interface {
	Compare(ref Tokens) (bool, []rune)
	IsSame(a, b []rune) bool
	CompareRune(a, b rune) bool
	fmt.Stringer
}

type Scanner interface {
	Scan(docs ...Document) Matches
}
