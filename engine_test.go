package place

import (
	"archive/zip"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// маленькие дампы в формате geonames, чтобы тесты шли без сети
const testCities = `9000001	Moscow	Moscow	Moskva	55.75222	37.61556	P	PPLC	RU		48				10000000		144	Europe/Moscow	2024-01-01
9000002	Mosul	Mosul		36.335	43.11889	P	PPLA	IQ		15				1600000		228	Asia/Baghdad	2024-01-01
9000003	New York City	New York City	NYC	40.71427	-74.00597	P	PPL	US		NY				8000000	10	57	America/New_York	2024-01-01
9000004	Newport	Newport		51.58774	-2.99835	P	PPLA2	GB		WLS				300000		12	Europe/London	2024-01-01
9000005	Saint Petersburg	Saint Petersburg		59.93863	30.31413	P	PPLA	RU		66				5000000		11	Europe/Moscow	2024-01-01
9000006	Paris	Paris		48.85341	2.3488	P	PPLC	FR		11				2100000		42	Europe/Paris	2024-01-01
9000007	Paris	Paris		33.66094	-95.55551	P	PPLA2	US		TX				25000	183	180	America/Chicago	2024-01-01
9000008	Tiny	Tiny		1	1	P	PPL	US		TX				5	1	1	America/Chicago	2024-01-01
`

const testNames = `1	9000001	ru	Москва						
2	9000001	en	Moscow						
3	9000001	fr	Moscou						
4	9000001	de	Moskau						
5	9000001	link	https://en.wikipedia.org/wiki/Moscow						
6	9000002	ru	Мосул						
7	9000003	en	New York	1					
8	9000003	en	New York City						
9	9000003	ru	Нью-Йорк						
10	9000003	en	Big Apple			1			
11	9000003	en	New Amsterdam				1		
12	9000005	ru	Санкт-Петербург						
13	9000005	fr	Saint-Pétersbourg						
14	9000006	ru	Париж						
15	1	ru	Чужой						
`

func writeZip(t testing.TB, name, entry, text string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z := zip.NewWriter(f)
	w, _ := z.Create(entry)
	io.WriteString(w, text)
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
}

func testOptions(t testing.TB) *options {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, dirSource), 0o755)
	writeZip(t, filepath.Join(dir, dirSource, "cities15000.zip"), "cities15000.txt", testCities)
	writeZip(t, filepath.Join(dir, dirSource, "alternateNamesV2.zip"), "alternateNamesV2.txt", testNames)
	return NewOptions().Langs("en", "ru", "fr").Volume(dir).City15000().Log(log.New(io.Discard, "", 0))
}

var testFields = []string{FieldName, FieldCountry, FieldPeople, "location"}

func find(t testing.TB, e *Engine, q, lang string, fields ...string) (res []map[string]any) {
	b := e.Search(10, q, lang, fields...)
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("%q: %v: %s", q, err, b)
	}
	return
}

func names(list []map[string]any) (res string) {
	var s []string
	for _, x := range list {
		s = append(s, x["name"].(string))
	}
	return strings.Join(s, "|")
}

func TestEngine(t *testing.T) {
	o := testOptions(t)
	e := New(o)

	if got := string(e.Search(10, "mos", "", testFields...)); got != `[{"country":"ru","id":9000001,"location":"Moscow, RU","name":"Moscow","people":10000000},`+
		`{"country":"iq","id":9000002,"location":"Mosul, IQ","name":"Mosul","people":1600000}]` {
		t.Error(got)
	}
	if got := string(e.Search(10, "zzz", "", testFields...)); got != `[]` {
		t.Error(got)
	}

	for _, c := range []struct{ q, lang, want string }{
		{"мос", "", "Москва|Мосул"},   // язык по письменности
		{"mos", "ru", "Москва|Мосул"}, // язык задан
		{"mos", "fr", "Moscou|Mosul"}, // нет перевода > основное название
		{"mos", "de", "Moscow|Mosul"}, // язык не разрешен > en
		{"moskau", "", ""},            // и по нему не ищем
		{"new york", "", "New York"},  // официальное название из двух en
		{"нью", "", "Нью-Йорк"},
		{"big", "", ""}, // разговорные пропускаем
		{"new", "", "New York|Newport"},
		{"pa", "", "Paris|Paris"},
		{"pa @fr", "", "Paris"},
		{"pa @tx", "", "Paris"},
		{"@us", "", "New York|Paris"},
		{"@uk", "", "Newport"}, // gb > uk
		{"saint petersburg", "", "Saint Petersburg"},
		{"санкт", "", "Санкт-Петербург"},
		{"tiny", "", ""}, // меньше 10 жителей
		{"чуж", "", ""},
	} {
		if got := names(find(t, e, c.q, c.lang, testFields...)); got != c.want {
			t.Errorf("%q %q: got %q, want %q", c.q, c.lang, got, c.want)
		}
	}

	if got := find(t, e, "pa @tx", "", "location")[0]["location"]; got != "Paris, TX" {
		t.Error(got)
	}
	if got := find(t, e, "нью", "", "location")[0]["location"]; got != "Нью-Йорк, NY" {
		t.Error(got)
	}

	// быстрая выдача из памяти обязана совпадать с полной записью из базы
	all := []string{FieldAlias, FieldBoost, FieldCountry, FieldCur, FieldFlag, FieldHash, FieldLang, FieldLat, FieldLong,
		FieldName, FieldPeople, FieldState, FieldSymbol, FieldTags, FieldTZ, FieldZip, "location"}
	for _, lang := range []string{"en", "ru", "fr"} {
		for _, x := range find(t, e, "@ru", lang, all...) {
			var want map[string]any
			json.Unmarshal(e.Get(int(x["id"].(float64)), lang, all...), &want)
			if !reflect.DeepEqual(x, want) {
				t.Errorf("fast != full:\n%v\n%v", x, want)
			}
		}
		for _, q := range []string{"n", "p", "@us"} {
			var want []map[string]any
			for _, x := range e.search.Search(10, q) {
				var m map[string]any
				json.Unmarshal(e.Get(x.ID, lang, all...), &m)
				want = append(want, m)
			}
			if got := find(t, e, q, lang, all...); !reflect.DeepEqual(got, want) {
				t.Errorf("%q %s fast != full:\n%v\n%v", q, lang, got, want)
			}
		}
	}

	// без полей и с names отдаются полные записи
	if x := find(t, e, "mos", "ru")[0]; x["origin"] != "Moscow" || x["name"] != "Москва" || x["names"] == nil {
		t.Error(x)
	}
	if x := find(t, e, "mos", "ru", FieldNames)[0]; x["names"].(map[string]any)["de"] != "Moskau" {
		t.Error(x)
	}

	// изменения
	e.Boost(9000002)
	if got := names(find(t, e, "mos", "", testFields...)); got != "Mosul|Moscow" {
		t.Error("boost", got)
	}
	e.RemoveBoost(9000002)
	if got := names(find(t, e, "mos", "", testFields...)); got != "Moscow|Mosul" {
		t.Error("remove boost", got)
	}
	e.UpdateNames(9000002, "en", "Nineveh")
	if got := names(find(t, e, "nin", "", testFields...)); got != "Nineveh" {
		t.Error("update names", got)
	}
	e.UpdateKey(9000004, FieldPeople, 9000000)
	if got := names(find(t, e, "new", "", testFields...)); got != "Newport|New York" {
		t.Error("update key", got)
	}
	e.Delete(e.Get(9000006, ""))
	if got := names(find(t, e, "pa", "", testFields...)); got != "Paris" {
		t.Error("delete", got)
	}

	// после перезапуска подключается готовый файл индекса, изменения на месте
	e.Close()
	e = New(o)
	defer e.Close()
	if !e.search.Loaded() {
		t.Error("index file is not reused")
	}
	for q, want := range map[string]string{"nin": "Nineveh", "new": "Newport|New York", "pa": "Paris", "мос": "Москва|Мосул"} {
		if got := names(find(t, e, q, "", testFields...)); got != want {
			t.Errorf("restart %q: got %q, want %q", q, got, want)
		}
	}

	// восстановление пачкой
	backup := e.Backup()
	e.UpdateKey(9000001, FieldName, "Moskva")
	if got := names(find(t, e, "moskva", "de", testFields...)); got != "Moscow" { // name ищется, en показывается
		t.Error("update name", got)
	}
	e.Restore(backup)
	if got := names(find(t, e, "nin", "", testFields...)); got != "Nineveh" {
		t.Error("restore", got)
	}
}

// без Close файл индекса устарел: изменения есть в базе, но не в нем
func TestCrash(t *testing.T) {
	o := testOptions(t)
	e := New(o)
	e.Close()

	e = New(o)
	if !e.search.Loaded() {
		t.Fatal("index file is not reused")
	}
	e.UpdateNames(9000002, "en", "Nineveh")
	e.Delete(e.Get(9000006, ""))
	e.db.Close() // падение: индекс не сброшен

	e = New(o)
	defer e.Close()
	if e.search.Loaded() {
		t.Fatal("stale index file is reused")
	}
	if got := names(find(t, e, "nin", "", testFields...)); got != "Nineveh" {
		t.Error(got)
	}
	if got := names(find(t, e, "pa", "", testFields...)); got != "Paris" {
		t.Error(got)
	}
}

func TestDetect(t *testing.T) {
	o := NewOptions().Langs("en", "ru", "ja")
	o.prepare()
	for q, want := range map[string]string{"": "en", "new": "en", "são": "en", "мос": "ru", "new йорк": "ru", "київ": "ru", "東京": "ja", "とう": "ja", "αθή": "en"} {
		if got := o.detect(q); got != want {
			t.Errorf("%q: got %q, want %q", q, got, want)
		}
	}
}

func TestSearchAllocs(t *testing.T) {
	e := New(testOptions(t))
	defer e.Close()
	// один буфер ответа и ничего больше
	if n := testing.AllocsPerRun(100, func() { e.Search(10, "mos", "", testFields...) }); n > 1 {
		t.Error("allocs", n)
	}
}
