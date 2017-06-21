package sunat

import "testing"

func Test_isName(t *testing.T) {
	fixtures := map[string]bool{
		"foo":      true,
		"foo bar":  true,
		"a":        false,
		"foo ":     false,
		"foo  bar": true,
		"x 1":      false,
		"ab":       true,
		"":         false,
	}
	for k, expected := range fixtures {
		if isName(k) != expected {
			t.Error("Failed isName with ", k, ". Expected: ", expected)
		}
	}
}

func Test_isRuc(t *testing.T) {
	fixtures := map[string]bool{
		"20254138577": true,
		"12345678909": false,
		"asdfghjkl12": false,
		"20601020841": true,
	}
	for k, expected := range fixtures {
		if isRuc(k) != expected {
			t.Error("Failed isRuc with ", k, ". Expected: ", expected)
		}
	}
}

func Test_isDni(t *testing.T) {
	fixtures := map[string]bool{
		"20254138577": false,
		"12345678909": false,
		"asdfghjkl12": false,
		"53140230":    true,
	}
	for k, expected := range fixtures {
		if isDni(k) != expected {
			t.Error("Failed isDni with ", k, ". Expected: ", expected)
		}
	}
}
