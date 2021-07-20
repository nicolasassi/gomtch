package document

import (
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"testing"
	"unicode"
)

func TestASCII_Transform(t *testing.T) {
	type fields struct {
		t transform.Transformer
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"default", fields{t: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)},
			args{s: ""}, "", false},
		{"café", fields{t: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)},
			args{s: "café"}, "cafe", false},
		{"randomAccentedChars", fields{t: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)},
			args{s: "éíóiü"}, "eioiu", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ASCII{
				t: tt.fields.t,
			}
			got, err := a.Transform(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Transform() got = %v, want %v", got, tt.want)
			}
		})
	}
}
