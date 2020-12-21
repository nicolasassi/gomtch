package gomtch

import "fmt"

type Document interface {
	Scan(docs ...Document) Matches
	Compare(ref Tokens) (bool, []rune)
	IsSame(a, b []rune) bool
	CompareRune(a, b rune) bool
	fmt.Stringer
}
