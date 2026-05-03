package words_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/words/words"
)

func TestNorm(t *testing.T) {
	cases := []struct {
		name     string
		phrase   string
		expected []string
	}{
		{
			name:     "empty phrase",
			phrase:   "",
			expected: []string{},
		},
		{
			name:     "stop words only",
			phrase:   "the a is are",
			expected: []string{},
		},
		{
			name:     "simple words",
			phrase:   "running cats dogs",
			expected: []string{"run", "cat", "dog"},
		},
		{
			name:     "duplicates removed",
			phrase:   "running runs run",
			expected: []string{"run"},
		},
		{
			name:     "punctuation ignored",
			phrase:   "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "mixed case",
			phrase:   "Hello WORLD",
			expected: []string{"hello", "world"},
		},
		{
			name:     "digits preserved",
			phrase:   "test123",
			expected: []string{"test123"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := words.Norm(tc.phrase)
			assert.ElementsMatch(t, tc.expected, result)
		})
	}
}