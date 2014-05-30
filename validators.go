package main

import (
	"regexp"
	"strconv"
)

func isRuc(value string) bool {
	if ok, _ := regexp.MatchString("^(10|15|17|20)\\d{9}$", value); !ok {
		return false
	}
	weights := [10]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	checkDigit, _ := strconv.Atoi(string(value[10]))
	sum := 0
	for i, weight := range weights {
		digit, _ := strconv.Atoi(string(value[i]))
		sum += weight * digit
	}
	return checkDigit == (11 - (sum % 11))
}

func isDni(value string) bool {
	ok, _ := regexp.MatchString("^\\d{8}$", value)
	return ok
}

func isName(value string) bool {
	ok, _ := regexp.MatchString("(?i)[^a-z\\s]", value)
	return !ok && len(value) > 4
}
