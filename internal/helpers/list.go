package helpers

import (
	"strings"
)

// UniqueOptions configures the behavior of UniqueStrings
type UniqueOptions struct {
	CaseSensitive bool
	PreserveOrder bool
}

// UniqueStrings removes duplicates from a slice of strings with configurable options
func UniqueStrings(slice []string, opts UniqueOptions) []string {
	if len(slice) == 0 {
		return slice
	}

	if opts.PreserveOrder {
		seen := make(map[string]bool)
		result := make([]string, 0, len(slice))

		for _, item := range slice {
			key := item
			if !opts.CaseSensitive {
				key = strings.ToLower(item)
			}

			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result
	} else {
		// Use map for O(1) lookup, order not preserved
		seen := make(map[string]bool)
		result := make([]string, 0, len(slice))

		for _, item := range slice {
			key := item
			if !opts.CaseSensitive {
				key = strings.ToLower(item)
			}

			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result
	}
}

// UniqueStringsSimple removes duplicates from a slice of strings (case-sensitive, order preserved)
func UniqueStringsSimple(slice []string) []string {
	return UniqueStrings(slice, UniqueOptions{
		CaseSensitive: true,
		PreserveOrder: true,
	})
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UniqueStringsByKey removes duplicates from a slice using a key function
func UniqueStringsByKey(items []string, keyFunc func(string) string, opts UniqueOptions) []string {
	return UniqueStructByKey(items, keyFunc, opts)
}

func UniqueStructByKey[T any](items []T, keyFunc func(T) string, opts UniqueOptions) []T {
	if len(items) == 0 {
		return items
	}

	seen := make(map[string]bool)
	result := make([]T, 0, len(items))

	for _, item := range items {
		key := keyFunc(item)
		if !opts.CaseSensitive && opts.PreserveOrder {
			key = strings.ToLower(key)
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}
