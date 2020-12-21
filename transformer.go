package gomtch

import (
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

type Transformer interface {
	Transform(s string) (string, error)
}

type ASCII struct {
	t transform.Transformer
}

func NewASCII() *ASCII {
	return &ASCII{t: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)}
}

func (a ASCII) Transform(s string) (string, error) {
	cntr := 1
	dst := make([]byte, len(s)*cntr)
	for {
		nDst, _, err := a.t.Transform(dst, []byte(s), true)
		if err != nil {
			if err == transform.ErrShortDst {
				cntr++
				dst = make([]byte, len(s)*cntr)
				continue
			}
			return "", err
		}
		return string(dst[:nDst]), nil
	}
}
