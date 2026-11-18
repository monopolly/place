package model

import (
	"bytes"
	"unicode"
)

// New York > NewYork
func clear(f string) (r string) {
	var b bytes.Buffer
	for _, letter := range f {
		if unicode.IsLetter(letter) {
			b.WriteRune(letter)
		}
	}
	return b.String()
}
