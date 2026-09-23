package strutil

import (
	"unicode"
	"unicode/utf8"
)

// FirstToLower converts the first character of a string to lowercase.
// Characters without a lowercase form ('_', digits, ...) are left untouched.
func FirstToLower(s string) string {
	if s == "" {
		return s
	}
	// ASCII fast path.
	if c := s[0]; c < utf8.RuneSelf {
		if 'A' <= c && c <= 'Z' {
			return string(c+'a'-'A') + s[1:]
		}
		return s
	}

	r, size := utf8.DecodeRuneInString(s)
	lower := unicode.ToLower(r)
	if lower == r {
		return s
	}
	return string(lower) + s[size:]
}
