package strutil

// FirstToLower converts the first character of a string to lowercase.
func FirstToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]|32) + s[1:]
}
