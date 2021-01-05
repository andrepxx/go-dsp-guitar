package path

import (
	"strings"
)

/*
 * Checks whether a string starts and ends with a quotation mark.
 */
func isQuoted(s string) bool {
	r := []rune(s)
	numChars := len(r)
	short := numChars < 2

	/*
	 * Check if string is too short.
	 */
	if short {
		return false
	} else {
		leadingSingle := strings.HasPrefix(s, "'")
		trailingSingle := strings.HasSuffix(s, "'")
		quotedSingle := leadingSingle && trailingSingle
		leadingDouble := strings.HasPrefix(s, "\"")
		trailingDouble := strings.HasSuffix(s, "\"")
		quotedDouble := leadingDouble && trailingDouble
		quoted := quotedSingle || quotedDouble
		return quoted
	}

}

/*
 * Sanitize path, removing leading and trailing quotes and white space.
 */
func Sanitize(p string) string {
	p = strings.TrimSpace(p)
	quoted := isQuoted(p)

	/*
	 * Remove quotation marks, if present.
	 */
	if quoted {
		r := []rune(p)
		numChars := len(r)
		lastIdx := numChars - 1
		r = r[1:lastIdx]
		p = string(r)
	}

	return p
}
