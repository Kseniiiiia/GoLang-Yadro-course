package words

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNorm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    " ",
			expected: []string(nil),
		},
		{
			name:     "simple words",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with punctuation",
			input:    "Hello, world! How are you?",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with stop words",
			input:    "the quick brown fox jumps over the lazy dog",
			expected: []string{"quick", "brown", "fox", "jump", "lazi", "dog"},
		},
		{
			name:     "with stemming",
			input:    "running jumps quickly",
			expected: []string{"run", "jump", "quick"},
		},
		{
			name:     "with plus signs",
			input:    "c++ programming+language",
			expected: []string{"c", "program", "languag"},
		},
		{
			name:     "with mixed case",
			input:    "GoLang PYTHON JavaScript",
			expected: []string{"golang", "python", "javascript"},
		},
		{
			name:     "duplicate words",
			input:    "test test test",
			expected: []string{"test"},
		},
		{
			name:     "only stop words",
			input:    "the and or but",
			expected: []string(nil),
		},
		{
			name:     "with special characters",
			input:    "email@gmail.com http://site.com",
			expected: []string{"email", "gmail", "com", "http", "site"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Norm(tt.input)

			sortStrings(result)
			sortStrings(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func sortStrings(s []string) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}
