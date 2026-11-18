package place

import (
	"fmt"
	"log"
	"sort"
	"unicode"
	"unicode/utf8"
)

func NewOptions() (a *options) {
	a = new(options)
	a.langmap = make(map[string]bool)
	a.Langs("en")
	a.log = log.Default()
	return
}

// // "lang":["en","ru"], "volume":"/nice", "changes":""
// func NewOptionsJson(b []byte) (a *options) {
// 	a = new(options)
// 	a.langmap = make(map[string]bool)
// 	a.Langs("en")
// 	a.volume = "db"
// 	a.log = log.Default()
// 	return
// }

type options struct {
	volume  string
	langmap map[string]bool
	source  source
	log     *log.Logger
	version string
	zip     bool

	// считаются в prepare, дальше только читаются
	langs     []string       // языки по алфавиту
	langIndex map[string]int // язык > номер в langs
	scripts   []scriptLang   // язык для каждой письменности
}

// prepare фиксирует настройки перед стартом движка
func (a *options) prepare() {
	a.langs = a.langs[:0]
	for x := range a.langmap {
		a.langs = append(a.langs, x)
	}
	sort.Strings(a.langs)
	a.langIndex = make(map[string]int, len(a.langs))
	for i, x := range a.langs {
		a.langIndex[x] = i
	}
	a.scripts = a.scripts[:0]
	for _, s := range scripts {
		for _, l := range s.langs {
			if a.langmap[l] {
				a.scripts = append(a.scripts, scriptLang{s.table, l})
				break
			}
		}
	}
}

type scriptLang struct {
	table *unicode.RangeTable
	lang  string
}

// письменности и языки, которые на них пишут, по убыванию приоритета
var scripts = []struct {
	table *unicode.RangeTable
	langs []string
}{
	{unicode.Cyrillic, []string{"ru", "uk", "be", "bg", "sr", "mk", "kk"}},
	{unicode.Greek, []string{"el"}},
	{unicode.Hebrew, []string{"he"}},
	{unicode.Arabic, []string{"ar", "fa", "ur"}},
	{unicode.Hiragana, []string{"ja"}},
	{unicode.Katakana, []string{"ja"}},
	{unicode.Han, []string{"zh", "ja"}},
	{unicode.Hangul, []string{"ko"}},
	{unicode.Thai, []string{"th"}},
	{unicode.Devanagari, []string{"hi"}},
	{unicode.Bengali, []string{"bn"}},
	{unicode.Georgian, []string{"ka"}},
	{unicode.Armenian, []string{"hy"}},
}

// detect определяет язык запроса по письменности первой нелатинской буквы.
// На двух-трех буквах префикса это надежнее статистики и почти бесплатно.
// Латиница всегда "en": по началу слова языки на латинице не различить.
func (a *options) detect(q string) string {
	for i := 0; i < len(q); i++ {
		if q[i] < utf8.RuneSelf {
			continue
		}
		r, _ := utf8.DecodeRuneInString(q[i:])
		for _, s := range a.scripts {
			if unicode.Is(s.table, r) {
				return s.lang
			}
		}
		break
	}
	return "en"
}

func (a *options) City500() *options {
	a.source = city500
	return a
}
func (a *options) City1000() *options {
	a.source = city1000
	return a
}

func (a *options) City5000() *options {
	a.source = city5000
	return a
}

func (a *options) City15000() *options {
	a.source = city15000
	return a
}

func (a *options) Volume(volume string) *options {
	a.volume = volume
	return a
}
func (a *options) Version(v string) *options {
	a.version = v
	return a
}
func (a *options) Log(v *log.Logger) *options {
	a.log = v
	return a
}

// options hash for reindex
func (a *options) hash() string {
	var list []string
	for x := range a.langmap {
		list = append(list, x)
	}
	sort.Strings(list)
	return fmt.Sprint(list, a.source.name(), a.volume, a.version)
}

func (a *options) Langs(lang ...string) *options {

	for _, x := range lang {
		a.langmap[x] = true
	}
	return a
}

func (a *options) haslang(lang string) (has bool) {
	return a.langmap[lang]
}

func (a *options) Zipcodes() {
	a.zip = true
}
