package sunat

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	dniPattern    = regexp.MustCompile(`^\d{8}$`)
	rucPattern    = regexp.MustCompile(`^(10|15|17|20)\d{9}$`)
	namePattern   = regexp.MustCompile(`(^\w+(\s+\w+)*\w$)`)
	spacesPattern = regexp.MustCompile(`\s+`)
	digitsPattern = regexp.MustCompile(`(\d+)`)
)

func isRuc(s string) bool {
	if !rucPattern.MatchString(s) {
		return false
	}
	weights := [10]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	checkDigit, _ := strconv.Atoi(string(s[10]))
	sum := 0
	for i, weight := range weights {
		digit, _ := strconv.Atoi(string(s[i]))
		sum += weight * digit
	}
	mod := 11 - (sum % 11)
	if mod >= 10 {
		mod -= 10
	}
	return checkDigit == mod
}

func isDni(s string) bool {
	return dniPattern.MatchString(s)
}

func isName(s string) bool {
	return namePattern.MatchString(s)
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func removeExtraSpaces(s string) string {
	return trim(spacesPattern.ReplaceAllLiteralString(s, " "))
}
