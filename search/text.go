package search

import (
	"bytes"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// больше слов в запросе не бывает, лишние отбрасываются
const maxWords = 8

// scratch это рабочая память одного запроса. Берется из пула, поэтому поиск
// ничего не выделяет (локальные буферы утекли бы в кучу через x/text/norm).
type scratch struct {
	buf, b1, b2 []byte
	words       [maxWords][]byte
	heap        []span
	items       []Item
}

var pool = sync.Pool{New: func() any { return new(scratch) }}

// normalize возвращает каноничный вид строки: нижний регистр, без диакритики,
// только буквы, цифры, ':' и '@' — все остальное становится пробелом.
// Один и тот же вид используется и для индекса, и для запроса.
// Результат живет в scratch до следующего вызова.
func (a *scratch) normalize(s string) []byte {
	dst := a.buf[:0]
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		for i := 0; i < len(s); i++ {
			c := s[i]
			switch {
			case 'a' <= c && c <= 'z', '0' <= c && c <= '9', c == ':', c == '@':
			case 'A' <= c && c <= 'Z':
				c += 'a' - 'A'
			default:
				c = ' '
			}
			dst = append(dst, c)
		}
		a.buf = dst
		return dst
	}

	// раскладываем на базовые буквы и знаки, знаки (Mn) выкидываем, собираем обратно
	a.b1 = norm.NFD.AppendString(a.b1[:0], s)
	d, t := a.b1, a.b2[:0]
	for len(d) > 0 {
		r, n := utf8.DecodeRune(d)
		d = d[n:]
		switch {
		case unicode.Is(unicode.Mn, r):
		case r == ':' || r == '@' || unicode.IsLetter(r) || unicode.IsDigit(r):
			t = utf8.AppendRune(t, unicode.ToLower(r))
		default:
			t = append(t, ' ')
		}
	}
	a.b2 = t
	a.buf = norm.NFC.Append(dst, t...)
	return a.buf
}

// пределы одного документа, под них упакованы ссылки на слова при сборке
const (
	maxWord = 1<<12 - 1
	maxText = 1<<20 - 1
)

// newSource готовит документ к индексации
func newSource(item Item, text string, options []string) *source {
	sc := pool.Get().(*scratch)
	defer pool.Put(sc)
	b := sc.normalize(text)

	// уникальные слова через пробел
	dst := sc.b1[:0]
	for i := 0; i < len(b); {
		if b[i] == ' ' {
			i++
			continue
		}
		j := i
		for j < len(b) && b[j] != ' ' {
			j++
		}
		w := b[i:j]
		i = j
		if len(w) > maxWord || len(dst)+len(w) >= maxText || hasWord(dst, w, ' ') {
			continue
		}
		if len(dst) > 0 {
			dst = append(dst, ' ')
		}
		dst = append(dst, w...)
	}
	split := len(dst)

	// уникальные опции через 0
	for _, x := range options {
		x = strings.ToLower(x)
		if x == "" || len(x) > maxWord || len(dst)+len(x) >= maxText || strings.IndexByte(x, 0) >= 0 || hasWord(dst[split:], []byte(x), 0) {
			continue
		}
		if len(dst) > split {
			dst = append(dst, 0)
		}
		dst = append(dst, x...)
	}
	sc.b1 = dst
	return &source{Item: item, text: string(dst), split: uint32(split)}
}

// hasWord ищет слово в списке через sep
func hasWord(list, w []byte, sep byte) bool {
	for len(list) > 0 {
		x := list
		if i := bytes.IndexByte(list, sep); i >= 0 {
			x, list = list[:i], list[i+1:]
		} else {
			list = nil
		}
		if string(x) == string(w) {
			return true
		}
	}
	return false
}

// next отрезает от списка через sep первое слово
func next(list string, sep byte) (word, rest string) {
	if i := strings.IndexByte(list, sep); i >= 0 {
		return list[:i], list[i+1:]
	}
	return list, ""
}

// parseQuery разбирает запрос на слова без аллокаций.
// Слова с '@' — отдельные фильтры, а идущие подряд обычные слова склеиваются
// в одно: "new york @us" > "newyork", "@us". Названия в индексе тоже склеены.
func (a *scratch) parseQuery(text string) [][]byte {
	b := a.normalize(text)
	words := a.words[:0]
	w := 0      // куда пишем склеенное слово
	start := -1 // его начало, -1 если слова нет
	for i := 0; i < len(b); {
		if b[i] == ' ' {
			i++
			continue
		}
		j := i
		for j < len(b) && b[j] != ' ' {
			j++
		}
		if b[i] == '@' {
			if start >= 0 {
				words = appendWord(words, b[start:w:w])
				start = -1
			}
			words = appendWord(words, b[i:j:j])
			w = j
		} else {
			if start < 0 {
				start = w
			}
			w += copy(b[w:], b[i:j])
		}
		i = j
	}
	if start >= 0 {
		words = appendWord(words, b[start:w:w])
	}
	return words
}

func appendWord(words [][]byte, w []byte) [][]byte {
	if len(words) == maxWords {
		return words
	}
	for _, x := range words {
		if string(x) == string(w) {
			return words
		}
	}
	return append(words, w)
}

func hasPrefix(s string, p []byte) bool {
	return len(s) >= len(p) && s[:len(p)] == string(p)
}
