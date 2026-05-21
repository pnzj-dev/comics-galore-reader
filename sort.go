package cgreaderwasm

import (
	"strconv"
	"unicode"
)

// NaturalLess compares two strings using natural sort order.
// "page_2.jpg" < "page_10.jpg" (lexicographic would put 10 before 2).
func NaturalLess(a, b string) bool {
	ra := []rune(a)
	rb := []rune(b)

	ia, ib := 0, 0
	for ia < len(ra) && ib < len(rb) {
		ca := ra[ia]
		cb := rb[ib]

		// If both are digits, parse the full number.
		if unicode.IsDigit(ca) && unicode.IsDigit(cb) {
			// Parse number in a.
			na, advA := parseNumber(ra[ia:])
			nb, advB := parseNumber(rb[ib:])

			if na != nb {
				return na < nb
			}
			ia += advA
			ib += advB
			continue
		}

		if ca != cb {
			return ca < cb
		}
		ia++
		ib++
	}

	// Shorter string comes first if all else equal.
	return len(ra)-ia < len(rb)-ib
}

// parseNumber extracts a leading integer from a rune slice.
func parseNumber(r []rune) (int, int) {
	end := 0
	for end < len(r) && unicode.IsDigit(r[end]) {
		end++
	}
	if end == 0 {
		return 0, 0
	}
	n, err := strconv.Atoi(string(r[:end]))
	if err != nil {
		return 0, 0
	}
	return n, end
}
