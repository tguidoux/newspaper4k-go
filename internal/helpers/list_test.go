package helpers

import (
	"reflect"
	"strings"
	"testing"
)

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		opts     UniqueOptions
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: true},
			expected: []string{},
		},
		{
			name:     "no duplicates, case sensitive, preserve order",
			input:    []string{"a", "b", "c"},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: true},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates, case sensitive, preserve order",
			input:    []string{"a", "b", "a", "c", "b"},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: true},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates, case insensitive, preserve order",
			input:    []string{"a", "B", "A", "c", "b"},
			opts:     UniqueOptions{CaseSensitive: false, PreserveOrder: true},
			expected: []string{"a", "B", "c"},
		},
		{
			name:     "with duplicates, case sensitive, no preserve order",
			input:    []string{"c", "b", "a", "c", "b"},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: false},
			expected: []string{"c", "b", "a"}, // order may vary, but we'll check length and contents
		},
		{
			name:     "with duplicates, case insensitive, no preserve order",
			input:    []string{"C", "b", "A", "c", "B"},
			opts:     UniqueOptions{CaseSensitive: false, PreserveOrder: false},
			expected: []string{"C", "b", "A"}, // order may vary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStrings(tt.input, tt.opts)
			if tt.opts.PreserveOrder {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("UniqueStrings() = %v, expected %v", result, tt.expected)
				}
			} else {
				// For non-preserving order, check length and that all expected items are present
				if len(result) != len(tt.expected) {
					t.Errorf("UniqueStrings() length = %d, expected %d", len(result), len(tt.expected))
				}
				// Check that all expected items are in result (considering case sensitivity)
				expectedMap := make(map[string]bool)
				for _, item := range tt.expected {
					key := item
					if !tt.opts.CaseSensitive {
						key = strings.ToLower(item)
					}
					expectedMap[key] = true
				}
				for _, item := range result {
					key := item
					if !tt.opts.CaseSensitive {
						key = strings.ToLower(item)
					}
					if !expectedMap[key] {
						t.Errorf("UniqueStrings() contains unexpected item %v", item)
					}
				}
			}
		})
	}
}

func TestUniqueStringsSimple(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "case sensitive duplicates",
			input:    []string{"a", "A", "a"},
			expected: []string{"a", "A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStringsSimple(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("UniqueStringsSimple() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{
			name:     "a less than b",
			a:        1,
			b:        2,
			expected: 1,
		},
		{
			name:     "b less than a",
			a:        5,
			b:        3,
			expected: 3,
		},
		{
			name:     "a equals b",
			a:        4,
			b:        4,
			expected: 4,
		},
		{
			name:     "negative numbers",
			a:        -1,
			b:        -2,
			expected: -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Min() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestUniqueStringsByKey(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		keyFunc  func(string) string
		opts     UniqueOptions
		expected []string
	}{
		{
			name:  "empty slice",
			input: []string{},
			keyFunc: func(s string) string {
				return s
			},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: true},
			expected: []string{},
		},
		{
			name:  "no duplicates by key",
			input: []string{"apple", "banana", "cherry"},
			keyFunc: func(s string) string {
				return s[:1] // first letter
			},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: true},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:  "with duplicates by key",
			input: []string{"apple", "avocado", "banana", "cherry"},
			keyFunc: func(s string) string {
				return s[:1] // first letter
			},
			opts:     UniqueOptions{CaseSensitive: true, PreserveOrder: true},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:  "case insensitive key",
			input: []string{"Apple", "avocado", "Banana", "cherry"},
			keyFunc: func(s string) string {
				return s[:1] // first letter
			},
			opts:     UniqueOptions{CaseSensitive: false, PreserveOrder: true},
			expected: []string{"Apple", "Banana", "cherry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStringsByKey(tt.input, tt.keyFunc, tt.opts)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("UniqueStringsByKey() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
