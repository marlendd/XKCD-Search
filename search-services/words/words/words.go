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
	condFunc := func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}

	words := strings.FieldsFunc(phrase, condFunc)
	seen := make(map[string]struct{})

	for _, word := range words {
		if word == "" {
			continue
		}

		word = strings.ToLower(word)

		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			return nil
		}

		if english.IsStopWord(stemmed) {
			continue
		}

		if _, ok := seen[stemmed]; !ok {
			seen[stemmed] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(seen))
}
