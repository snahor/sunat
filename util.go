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

// mod11 validation will be used only for numbers starting with 10
func isRuc(s string) bool {
	if !rucPattern.MatchString(s) {
		return false
	}
	if s[:2] != "10" {
		return true
	}
	weights := [10]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	checkDigit, _ := strconv.Atoi(string(s[10]))
	sum := 0
	for i, weight := range weights {
		digit, _ := strconv.Atoi(string(s[i]))
		sum += weight * digit
	}
	return checkDigit == (11 - (sum % 11))
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
