package utils

import (
	"unicode/utf8"
)

func TruncateMiddle(s string, maxLen int) string {
	// 1. Check if truncation is even necessary
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	// 2. Handle edge case: maxLen is too small for "..."
	if maxLen <= 3 {
		return "..."
	}

	runes := []rune(s)
	// Calculate how many characters to keep on each side
	// (maxLen - 3 for the dots)
	sideLen := (maxLen - 3) / 2

	start := string(runes[:sideLen])
	end := string(runes[len(runes)-sideLen:])

	return start + "..." + end
}
