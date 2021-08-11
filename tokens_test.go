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
			Values: map[int][]rune{
				0: []rune("something"),
			},
			Ids: []int{0},
		}},
		{"twoWords", args{tokens: []string{"something", "else"}}, Tokens{
			Values: map[int][]rune{
				0: []rune("something"),
				1: []rune("else"),
			},
			Ids: []int{0, 1},
		}},
		{"duplicatedWord", args{tokens: []string{"something", "else", "something"}}, Tokens{
			Values: map[int][]rune{
				0: []rune("something"),
				1: []rune("else"),
			},
			Ids: []int{0, 1, 0},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMappingFromTokens(tt.args.tokens).Map()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokens_mapTokens(t1 *testing.T) {
	type args struct {
		m Mapping
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"default", args{
			m: NewMappingFromTokens([]string{"a", "b", "a", "c", "c", "a", "b", "d"}),
		},
			[]int{0, 1, 0, 3, 3, 0, 1, 7}},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			ids := tt.args.m.Map().Ids
			if !reflect.DeepEqual(tt.want, ids) {
				t1.Errorf("mapTokens() = %v, want %v", ids, tt.want)
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
		want Mapping
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
			if got := NewMappingFromTokens(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMappingFromTokens() = %v, want %v", got, tt.want)
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
