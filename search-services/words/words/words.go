package words

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)

func Norm(phrase string) []string {

	f := func(c rune) bool {
		return unicode.IsPunct(c) || unicode.IsSpace(c) || c == '+'
	}

	split := strings.FieldsFunc(phrase, f)
	words := make(map[string]bool)

	for _, word := range split {
		stemmed, _ := snowball.Stem(word, "english", false)

		if english.IsStopWord(stemmed) {
			continue
		}
		words[stemmed] = true
	}

	return slices.Collect(maps.Keys(words))
}
