package gomtch

import (
	"reflect"
	"testing"
)

func TestMakeTokens(t *testing.T) {
	type args struct {
		tokens []string
	}
	tests := []struct {
		name string
		args args
		want Tokens
	}{
		{"default", args{tokens: []string{"something"}}, Tokens{
			values: map[int][]rune{
				0: []rune("something"),
			},
			mapping: []int{0},
		}},
		{"twoWords", args{tokens: []string{"something", "else"}}, Tokens{
			values: map[int][]rune{
				0: []rune("something"),
				1: []rune("else"),
			},
			mapping: []int{0, 1},
		}},
		{"duplicatedWord", args{tokens: []string{"something", "else", "something"}}, Tokens{
			values: map[int][]rune{
				0: []rune("something"),
				1: []rune("else"),
			},
			mapping: []int{0, 1, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeTokens(tt.args.tokens)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokens_mapTokens(t1 *testing.T) {
	type args struct {
		m mapping
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"default", args{
			m: makeMapping([]string{"a", "b", "a", "c", "c", "a", "b", "d"}),
		},
			[]int{0, 1, 0, 3, 3, 0, 1, 7}},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tokens{
				values: map[int][]rune{},
			}
			t.mapTokens(tt.args.m)
			if !reflect.DeepEqual(tt.want, t.mapping) {
				t1.Errorf("mapTokens() = %v, want %v", t.mapping, tt.want)
			}
		})
	}
}

func Test_makeMapping(t *testing.T) {
	type args struct {
		v []string
	}
	tests := []struct {
		name string
		args args
		want mapping
	}{
		{"default0", args{v: []string{"comida", "gostosa"}}, map[string][]int{
			"comida":  {0},
			"gostosa": {1},
		}},
		{"duplicated", args{v: []string{"boa", "comida", "boa", "comida"}}, map[string][]int{
			"boa":    {0, 2},
			"comida": {1, 3},
		}},
		{"default1", args{v: []string{"comida", "gostosa."}}, map[string][]int{
			"comida":  {0},
			"gostosa": {1},
			".":       {2},
		}},
		{"default2", args{v: []string{"comida", ".gostosa"}}, map[string][]int{
			"comida":  {0},
			".":       {1},
			"gostosa": {2},
		}},
		{"default3", args{v: []string{"comida:", ".gostosa"}}, map[string][]int{
			"comida":  {0},
			":":       {1},
			".":       {2},
			"gostosa": {3},
		}},
		{"default4", args{v: []string{":comida:", ".gostosa"}}, map[string][]int{
			":":       {0, 2},
			"comida":  {1},
			".":       {3},
			"gostosa": {4},
		}},
		{"default5", args{v: []string{":comida", ".gostosa"}}, map[string][]int{
			":":       {0},
			"comida":  {1},
			".":       {2},
			"gostosa": {3},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeMapping(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEndSpecial(t *testing.T) {
	type args struct {
		token []rune
	}
	tests := []struct {
		name string
		args args
		want []rune
	}{
		{"default", args{token: []rune("amigo")}, nil},
		{"endWithDot", args{token: []rune("amigo.")}, []rune(".")},
		{"endWithManyDots", args{token: []rune("amigo...")}, []rune("...")},
		{"endWithDotAndHasSpecialInside", args{token: []rune("amig.o.")}, []rune(".")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEndSpecial(tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEndSpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStartSpecial(t *testing.T) {
	type args struct {
		token []rune
	}
	tests := []struct {
		name string
		args args
		want []rune
	}{
		{"default", args{token: []rune("amigo")}, nil},
		{"startWithDot", args{token: []rune(".amigo")}, []rune(".")},
		{"startWithManyDots", args{token: []rune("...amigo")}, []rune("...")},
		{"startWithDotAndHasSpecialInside", args{token: []rune(".a.migo")}, []rune(".")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStartSpecial(tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStartSpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}
