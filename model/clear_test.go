package model

import "testing"

func TestClear(t *testing.T) {
	for in, want := range map[string]string{
		"Rostov-On-Don": "RostovOnDon",
		"New York":      "NewYork",
		"Нью-Йорк 2":    "НьюЙорк",
		"":              "",
	} {
		if got := clear(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
