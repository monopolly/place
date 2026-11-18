// Package geo это справочники по странам и geohash, которые нужны при разборе
// geonames и при выдаче: флаг, валюта и язык страны, хеш координат.
package geo

// Country это данные страны
type Country struct {
	Flag   string // 🇷🇺
	Cur    string // RUB
	Symbol string // ₽
	Lang   string // ru, основной язык
}

// Lookup возвращает данные страны по коду в нижнем регистре: "ru", "us", "uk"
func Lookup(code string) Country {
	return countries[code]
}

const (
	base32    = "0123456789bcdefghjkmnpqrstuvwxyz"
	precision = 7 // ± 76 м
)

// Geohash кодирует координаты в 7 символов
func Geohash(lat, lon float64) string {
	var b [precision]byte
	return string(AppendGeohash(b[:0], lat, lon))
}

// AppendGeohash дописывает geohash в dst
func AppendGeohash(dst []byte, lat, lon float64) []byte {
	latMin, latMax := -90.0, 90.0
	lonMin, lonMax := -180.0, 180.0
	var idx, bit int
	even := true
	for n := 0; n < precision; {
		if even {
			if mid := (lonMin + lonMax) / 2; lon > mid {
				idx = idx*2 + 1
				lonMin = mid
			} else {
				idx = idx * 2
				lonMax = mid
			}
		} else {
			if mid := (latMin + latMax) / 2; lat > mid {
				idx = idx*2 + 1
				latMin = mid
			} else {
				idx = idx * 2
				latMax = mid
			}
		}
		even = !even
		if bit++; bit == 5 {
			dst = append(dst, base32[idx])
			bit, idx = 0, 0
			n++
		}
	}
	return dst
}
