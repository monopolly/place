package geo

import "testing"

func TestGeo(t *testing.T) {
	if got := Geohash(-23.5475, -46.63611); got != "6gyf4c2" { // Сан-Паулу
		t.Error(got)
	}
	if c := Lookup("ru"); c != (Country{"🇷🇺", "RUB", "₽", "ru"}) {
		t.Error(c)
	}
	if c := Lookup("uk"); c.Flag != "🇬🇧" || c.Cur != "GBP" {
		t.Error(c)
	}
	if c := Lookup("zz"); c != (Country{}) {
		t.Error(c)
	}
}
