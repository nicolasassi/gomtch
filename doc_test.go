package gomtch

import (
	"fmt"
	"github.com/jdkato/prose/tokenize"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNewDoc(t *testing.T) {
	type args struct {
		text io.Reader
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Doc
		wantErr bool
	}{
		{"default", args{text: strings.NewReader("cocaina")}, &Doc{
			Text:   "cocaina",
			Tokens: strings.Split("cocaina", " "),
		}, false},
		{"withHMTLParsingPTag", args{
			text: strings.NewReader("<p>cocaina</p>"), opts: []Option{WithHMTLParsing()}}, &Doc{
			Text:   "cocaina",
			Tokens: []string{"cocaina"},
		}, false},
		{"withHMTLParsingBTag", args{
			text: strings.NewReader("<b>cocaina</b>"), opts: []Option{WithHMTLParsing()}}, &Doc{
			Text:   "cocaina",
			Tokens: []string{"cocaina"},
		}, false},
		{"withTransform", args{
			text: strings.NewReader("cocaína"), opts: []Option{WithTransform(NewASCII())}}, &Doc{
			Text:   "cocaina",
			Tokens: []string{"cocaina"},
		}, false},
		{"withSequentialEqualCharsRemoval", args{
			text: strings.NewReader("cocaiiiina"), opts: []Option{WithSequentialEqualCharsRemoval()}},
			&Doc{
				Text:   "cocaina",
				Tokens: []string{"cocaina"},
			}, false},
		{"withSequentialEqualCharsRemoval", args{
			text: strings.NewReader("cocaííína"), opts: []Option{WithSequentialEqualCharsRemoval()}},
			&Doc{
				Text:   "cocaína",
				Tokens: []string{"cocaína"},
			}, false},
		{"withSequentialEqualCharsRemoval", args{
			text: strings.NewReader("iphone 11"), opts: []Option{WithSequentialEqualCharsRemoval()}},
			&Doc{
				Text:   "iphone 11",
				Tokens: []string{"iphone", "11"},
			}, false},
		{"withSetLower", args{
			text: strings.NewReader("Cocaína"), opts: []Option{WithSetLower()}},
			&Doc{
				Text:   "cocaína",
				Tokens: []string{"cocaína"},
			}, false},
		{"withSetUpper", args{
			text: strings.NewReader("Cocaína"), opts: []Option{WithSetUpper()}},
			&Doc{
				Text:   "COCAÍNA",
				Tokens: []string{"COCAÍNA"},
			}, false},
		{"withReplacer", args{
			text: strings.NewReader("coca (cocaína) para compra-venda"),
			opts: []Option{WithReplacer(regexp.MustCompile(`[()-]`), " ")}},
			&Doc{
				Text:   "coca  cocaína  para compra venda",
				Tokens: []string{"coca", "", "cocaína", "", "para", "compra", "venda"},
			}, false},
		{"withCustomRegexpTokenizer", args{
			text: strings.NewReader("coca cocaína para compra venda"),
			opts: []Option{}},
			&Doc{
				Text:   "coca cocaína para compra venda",
				Tokens: []string{"coca", "cocaína", "para", "compra", "venda"},
			}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDoc(tt.args.text, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("NewDoc() got = nil")
				return
			}
			if !reflect.DeepEqual(got.Tokens, tt.want.Tokens) ||
				!reflect.DeepEqual(got.Text, tt.want.Text) ||
				!reflect.DeepEqual(got.optError, tt.want.optError) ||
				!reflect.DeepEqual(got.transformer, tt.want.transformer) {
				t.Errorf("NewDoc() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDoc_String(t *testing.T) {
	type fields struct {
		t        transform.Transformer
		optError error
		text     string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"default", fields{
			text: "some Text",
		}, "some Text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Doc{
				transformer: tt.fields.t,
				optError:    tt.fields.optError,
				Text:        tt.fields.text,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoc_CompareRune(t *testing.T) {
	type args struct {
		a rune
		b rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"equalLetters", args{
			a: decodeRune("s"),
			b: decodeRune("s"),
		}, true},
		{"numbersDiff", args{
			a: decodeRune("1"),
			b: decodeRune("2"),
		}, false},
		{"numbersSame", args{
			a: decodeRune("1"),
			b: decodeRune("1"),
		}, true},
		{"numberAndNumericalInfo", args{
			a: decodeRune("1"),
			b: decodeRune("x"),
		}, false},
		{"punctAndNumber", args{
			a: decodeRune("."),
			b: decodeRune("1"),
		}, false},
		{"letterAndNumber", args{
			a: decodeRune("a"),
			b: decodeRune("1"),
		}, false},
		{"NumberAndletter", args{
			a: decodeRune("1"),
			b: decodeRune("a"),
		}, false},
		{"punctAndLetter", args{
			a: decodeRune("."),
			b: decodeRune("a"),
		}, true},
		{"letterAndPunct", args{
			a: decodeRune("i"),
			b: decodeRune("!"),
		}, false},
		{"punctAndPunct", args{
			a: decodeRune("!"),
			b: decodeRune("@"),
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Doc{}
			if got := d.CompareRune(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("CompareRune() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoc_IsSame(t *testing.T) {
	type fields struct {
		matchScoreFunc func(int, int) bool
		t              transform.Transformer
		optError       error
		text           string
	}
	type args struct {
		a []rune
		b []rune
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"default", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune("cocaina"),
			b: []rune("cocaina"),
		}, true},
		{"numberInCompared", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune("coc4ina"),
			b: []rune("cocaina"),
		}, true},
		{"numberInReference", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune("cocaina"),
			b: []rune("coc4ina"),
		}, false},
		{"allATo@InCompared", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune("c0c@in@"),
			b: []rune("cocaina"),
		}, false},
		{"smallWord", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune("at4"),
			b: []rune("ata"),
		}, true},
		{"pia", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune("pia"),
			b: []rune("pi1"),
		}, false},
		{"anal", fields{matchScoreFunc: func(matchScore, wordLength int) bool {
			return matchScore*100/wordLength >= 60
		}}, args{
			a: []rune(":)a="),
			b: []rune("anal"),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Doc{
				matchScoreFunc: tt.fields.matchScoreFunc,
				transformer:    tt.fields.t,
				optError:       tt.fields.optError,
				Text:           tt.fields.text,
			}
			if got := d.IsSame(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsSame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoc_Compare(t *testing.T) {
	type fields struct {
		text string
		opts []Option
	}
	type args struct {
		tokens Tokens
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		match    bool
		sequence []rune
	}{
		{"default", fields{
			text: "cocaína",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"cocaina"})},
			true,
			[]rune("cocaina"),
		},
		{"spacedWord", fields{
			text: "cocaína",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"co", "ca", "ina"})},
			true,
			[]rune("co ca ina"),
		},
		{"allSpacedWord", fields{
			text: "cocaína",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"c", "o", "c", "a", "i", "n", "a"})},
			true,
			[]rune("c o c a i n a"),
		},
		{"allSpacedWord", fields{
			text: "cocaína",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"co", "ca", "i", "na"})},
			true,
			[]rune("co ca i na"),
		},
		{"bigText", fields{
			text: "cocaína",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"abc", "ced", "cocaina", "a", "i", "n", "a"})},
			true,
			[]rune("cocaina"),
		},
		{"bigTextSplited", fields{
			text: "cocaína",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"abc", "ced", "coca", "ina", "a", "i", "n", "a"})},
			true,
			[]rune("coca ina"),
		},
		{"twoWords", fields{
			text: "cocaína branca",
			opts: []Option{WithTransform(
				NewASCII()), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil)}}, args{
			tokens: MakeTokens([]string{"cocaina", "branca"})},
			true,
			[]rune("cocaina branca"),
		},
		{"twoWordsSplited", fields{
			text: "cocaína branca",
			opts: []Option{
				WithTransform(NewASCII()), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil),
			}}, args{
			tokens: MakeTokens([]string{"cocaina", "br", "anca"})},
			true,
			[]rune("cocaina br anca"),
		},
		{"twoWordsAllSplited", fields{
			text: "cocaína branca",
			opts: []Option{
				WithTransform(NewASCII()), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil),
			}}, args{
			tokens: MakeTokens([]string{"c", "o", "c", "a", "i", "n", "a", "b", "r", "a", "n", "c", "a"})},
			true,
			[]rune("c o c a i n a b r a n c a"),
		},
		{"un! lever", fields{
			text: "unilever",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"un", "!", "lever"})},
			true,
			[]rune("un ! lever"),
		},
		{"uni lever", fields{
			text: "unilever",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"uni", "lever"})},
			true,
			[]rune("uni lever"),
		},
		{"specialBefore", fields{
			text: "unilever",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{".", "uni", "lever"})},
			true,
			[]rune("uni lever"),
		},
		{"pera", fields{
			text: "unilever",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"pera"})},
			false,
			nil,
		},
		{"pera", fields{
			text: "unilever bolada",
			opts: []Option{WithTransform(NewASCII()), WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{"unilever muito bolada"})},
			false,
			nil,
		},
		{"pera", fields{
			text: "natura",
			opts: []Option{WithTransform(NewASCII()), WithSetLower(),
				WithMinimumMatchScore(60)}}, args{
			tokens: MakeTokens([]string{`
										Ingredientes: Açúcar, água, suco concentrado cocaina de cassis e outras frutas,
										Womax aroma natural. Suco de fruta total: 29 % dos quais 23 % de cassis.
										<br>Não contém Glúten<br>Garrafa de Vidro<br>Cassis apresenta um sabor doce
										e levemente amargo, com uma cor escura. é tradicionalmente utilizado em
										geleias, sucos, sorvetes e xaropes.
									`})},
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDoc(strings.NewReader(tt.fields.text), tt.fields.opts...)
			if err != nil {
				t.Error(err)
			}
			if got, sequence := d.Compare(tt.args.tokens); got != tt.match {
				t.Errorf("Compare() = %v, want %v", got, tt.match)
			} else {
				if !reflect.DeepEqual(sequence, tt.sequence) {
					t.Log(sequence, tt.sequence)
					t.Errorf("Compare() = sequence: %v, want %v", string(sequence), string(tt.sequence))
				}
			}
		})
	}
}

func TestDoc_Scan(t *testing.T) {
	type args struct {
		docs []Document
	}
	tests := []struct {
		name string
		doc  *Doc
		args args
		want Matches
	}{
		{"default", func() *Doc {
			d, err := NewDoc(strings.NewReader(`Ingredientes: Açúcar, água, suco
			concentrado cocaina de cassis e outras frutas, Womax aroma natural. Suco de fruta
			total: 29 % dos quais 23 % de cassis.<br>Não contém Glúten<br>Garrafa de Vidro<br>Cassis
			apresenta um sabor doce e levemente amargo, com uma cor escura. é tradicionalmente utilizado
			em geleias, sucos, sorvetes e xaropes.`), WithTransform(NewASCII()), WithSetLower(),
				WithSequentialEqualCharsRemoval(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("Natura"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{}},
		{"cocaina", func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaína"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaína"), WithTransform(
				NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("cocaina"),
			}},
		{"conditionalMatchScoreMatch", func() *Doc {
			d, err := NewDoc(strings.NewReader("apple"), WithTransform(NewASCII()),
				WithSetLower(), WithConditionalMatchScore(matchScoreFunction), WithSequentialEqualCharsRemoval())
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("apple"), WithTransform(
				NewASCII()), WithSetLower(), WithConditionalMatchScore(matchScoreFunction),
				WithSequentialEqualCharsRemoval())
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("aple"),
			}},
		{"conditionalMatchScoreNotMatch", func() *Doc {
			d, err := NewDoc(strings.NewReader("g92"), WithTransform(NewASCII()),
				WithSetLower(), WithConditionalMatchScore(matchScoreFunction), WithSequentialEqualCharsRemoval())
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("*29"), WithTransform(
				NewASCII()), WithSetLower(), WithConditionalMatchScore(matchScoreFunction),
				WithSequentialEqualCharsRemoval())
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{}},
		{"defaultDifferentPuncts", func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaína"), WithTransform(
				NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaina"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("cocaina"),
			}},
		{"surroundedByDots", func() *Doc {
			d, err := NewDoc(strings.NewReader(".cocaína."), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaina"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("cocaina"),
			}},
		{"smallText", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s atibaia", string(b)))
			d, err := NewDoc(s,
				WithTransform(NewASCII()),
				WithSetLower(),
				WithMinimumMatchScore(60),
				WithCustomRegexpTokenizer(tokenize.NewRegexpTokenizer(`\s`, true, true)))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("atibaia"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("atibaia"),
			}},
		{"smallTextSep", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s ati baia", string(b)))
			d, err := NewDoc(s,
				WithTransform(NewASCII()),
				WithSetLower(),
				WithMinimumMatchScore(60),
				WithCustomRegexpTokenizer(tokenize.NewRegexpTokenizer(`\s`, true, true)))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("atibaia"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("ati baia"),
			}},
		{"smallTextSpecialSep", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s at! báia", string(b)))
			d, err := NewDoc(s,
				WithTransform(NewASCII()),
				WithSetLower(),
				WithMinimumMatchScore(60),
				WithCustomRegexpTokenizer(tokenize.NewRegexpTokenizer(`\s`, true, true)))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("atibaia"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("at ! baia"),
			}},
		{"smallTextManyWords", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s atibaia boa vida", string(b)))
			d, err := NewDoc(s,
				WithTransform(NewASCII()),
				WithSetLower(),
				WithMinimumMatchScore(60),
				WithCustomRegexpTokenizer(tokenize.NewRegexpTokenizer(`\s`, true, true)))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("atibaia boa vida"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}},
			map[int][]rune{
				0: []rune("atibaia boa vida"),
			}},
		{"smallTextManyDocs", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s atibaia boa vida Apple store", string(b)))
			d, err := NewDoc(s,
				WithTransform(NewASCII()),
				WithSetLower(),
				WithMinimumMatchScore(60),
				WithCustomRegexpTokenizer(tokenize.NewRegexpTokenizer(`\s`, true, true)))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{
			func() *Doc {
				d, err := NewDoc(strings.NewReader("atibaia vida"), WithTransform(NewASCII()),
					WithSetLower(), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil))
				if err != nil {
					log.Fatal(err)
				}
				return d
			}(),
			func() *Doc {
				d, err := NewDoc(strings.NewReader("Apple Store"), WithTransform(NewASCII()),
					WithSetLower(), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil))
				if err != nil {
					log.Fatal(err)
				}
				return d
			}(),
			func() *Doc {
				d, err := NewDoc(strings.NewReader("play store"), WithTransform(NewASCII()),
					WithSetLower(), WithMinimumMatchScore(60), WithCustomRegexpTokenizer(nil))
				if err != nil {
					log.Fatal(err)
				}
				return d
			}()}}, map[int][]rune{
			1: []rune("apple store"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.doc.Scan(tt.args.docs...); !reflect.DeepEqual(got, tt.want) {
				for k, v := range got {
					if vv, ok := tt.want[k]; ok {
						t.Logf("Scan() = key: %v got %v, want %v", k, string(v), string(vv))
						t.Logf("Scan() = key: %v got %v, want %v", k, v, vv)
					} else {
						t.Logf("Scan() = key: %v not found", k)
					}
				}
				t.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkDoc_Scan(b *testing.B) {
	type args struct {
		docs []Document
	}
	tests := []struct {
		name string
		doc  *Doc
		args args
		want Matches
	}{
		{"default", func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaína"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaína"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("cocaina"),
		}},
		{"defaultDifferentPuncts", func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaína"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaina"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("cocaina"),
		}},
		{"surroundedByDots", func() *Doc {
			d, err := NewDoc(strings.NewReader(".cocaína."), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("cocaina"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("cocaina"),
		}},
		{"smallText", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s atibaia", string(b)))
			d, err := NewDoc(s, WithTransform(NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("atibaia"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("atibaia"),
		}},
		{"smallTextSep", func() *Doc {
			f, err := os.Open("testdata/small_text.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s at! baia boa vida Apple store", string(b)))
			d, err := NewDoc(s, WithTransform(NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("atibaia"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("at!baia"),
		}},
		{"sertoes", func() *Doc {
			f, err := os.Open("testdata/sertoes.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s Unilever", string(b)))
			d, err := NewDoc(s, WithTransform(NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("Unilever"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("unilever"),
		}},
		{"sertoesSep", func() *Doc {
			f, err := os.Open("testdata/sertoes.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s Uni le ver", string(b)))

			d, err := NewDoc(s, WithTransform(NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("Unilever"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("unilever"),
		}},
		{"sertoesSpecialSep", func() *Doc {
			f, err := os.Open("testdata/sertoes.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s Un! le ver", string(b)))
			d, err := NewDoc(s, WithTransform(NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{func() *Doc {
			d, err := NewDoc(strings.NewReader("Unilever"), WithTransform(NewASCII()),
				WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}()}}, map[int][]rune{
			0: []rune("un!lever"),
		}},
		{"sertoesManyDocs", func() *Doc {
			f, err := os.Open("testdata/sertoes.txt")
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(f)
			s := strings.NewReader(fmt.Sprintf("%s Un! le ver a m i g o peixe urbano presente", string(b)))
			d, err := NewDoc(s, WithTransform(NewASCII()), WithSetLower(), WithMinimumMatchScore(60))
			if err != nil {
				log.Fatal(err)
			}
			return d
		}(), args{docs: []Document{
			func() *Doc {
				d, err := NewDoc(strings.NewReader("Unilever"), WithTransform(NewASCII()),
					WithSetLower(), WithMinimumMatchScore(60))
				if err != nil {
					log.Fatal(err)
				}
				return d
			}(),
			func() *Doc {
				d, err := NewDoc(strings.NewReader("amigo"), WithTransform(NewASCII()),
					WithSetLower(), WithMinimumMatchScore(60))
				if err != nil {
					log.Fatal(err)
				}
				return d
			}(),
			func() *Doc {
				d, err := NewDoc(strings.NewReader("peixe presente"), WithTransform(NewASCII()),
					WithSetLower(), WithMinimumMatchScore(60))
				if err != nil {
					log.Fatal(err)
				}
				return d
			}(),
		}}, map[int][]rune{
			0: []rune("un!lever"),
			1: []rune("amigo"),
		}},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			if got := tt.doc.Scan(tt.args.docs...); !reflect.DeepEqual(got, tt.want) {
				b.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func decodeRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

func matchScoreFunction(matchScore, wordLength int) bool {
	switch {
	case wordLength <= 4:
		return matchScore*100/wordLength == 100
	default:
		return matchScore*100/wordLength >= 60
	}
}
