package search

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func ids(list []Item) (res []int) {
	for _, x := range list {
		res = append(res, x.ID)
	}
	return
}

func TestSearch(t *testing.T) {
	s := New()
	s.Add(1, 100, 10, "NewYork НьюЙорк @dr5reg", []string{"@us", "@NY"})
	s.Add(2, 300, 0, "NewDelhi @ttnfv2", []string{"@in"})
	s.Add(3, 200, 0, "Newport São-Paulo @gcjwyz", []string{"@uk", "90210"})
	s.Add(4, 50, 20, "Москва moscow", []string{"@ru"})

	for _, c := range []struct {
		q    string
		want []int
	}{
		{"new", []int{1, 2, 3}}, // boost выше населения
		{"NEW", []int{1, 2, 3}},
		{"new york", []int{1}}, // слова склеиваются
		{"new-york", []int{1}},
		{"нью", []int{1}},
		{"нью иорк", []int{1}}, // й > и
		{"york", nil},          // только с начала названия
		{"new @us", []int{1}},
		{"@ny", []int{1}},
		{"@n", nil}, // опции только целиком
		{"new @uk", []int{3}},
		{"@uk new", []int{3}},
		{"new @us @ny @us", []int{1}},
		{"new @us @uk", nil},
		{"90210", []int{3}},
		{"9021", nil},
		{"sao", []int{3}}, // без диакритики
		{"são", []int{3}},
		{"@gcj", []int{3}}, // геохеш по префиксу
		{"м", []int{4}},
		{"m", []int{4}},
		{"", nil},
		{"  ,. ", nil},
		{"x", nil},
	} {
		if got := ids(s.Search(10, c.q)); !slices.Equal(got, c.want) {
			t.Errorf("%q: got %v, want %v", c.q, got, c.want)
		}
	}

	if got := ids(s.Search(2, "new")); !slices.Equal(got, []int{1, 2}) {
		t.Error("limit", got)
	}

	// замена и удаление
	s.Add(2, 300, 30, "NewDelhi", []string{"@in"})
	if got := ids(s.Search(10, "new")); !slices.Equal(got, []int{2, 1, 3}) {
		t.Error("update", got)
	}
	s.Delete(1)
	if got := ids(s.Search(10, "new")); !slices.Equal(got, []int{2, 3}) {
		t.Error("delete", got)
	}
	if s.Count() != 3 {
		t.Error("count", s.Count())
	}
}

// наивная реализация для сверки
type naive map[int]*source

func (a naive) search(limit int, text string) (res []int) {
	words := new(scratch).parseQuery(text)
	if len(words) == 0 {
		return
	}
	var list []Item
	for _, s := range a {
		if s.match(words) {
			list = append(list, s.Item)
		}
	}
	slices.SortFunc(list, func(x, y Item) int { return compare(&x, &y) })
	if len(list) > limit {
		list = list[:limit]
	}
	return ids(list)
}

func randWord(r *rand.Rand, abc string, max int) string {
	b := make([]rune, 1+r.Intn(max))
	letters := []rune(abc)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// случайные добавления, замены и удаления: поиск обязан совпадать с перебором
// и в дельте, и после пересборки, и внутри пакетной загрузки
func TestRandom(t *testing.T) {
	t.Run("memory", func(t *testing.T) { testRandom(t, New()) })
	t.Run("file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "search.idx")
		s := Open(path, 1)
		if s.Loaded() {
			t.Fatal("loaded")
		}
		ref := testRandom(t, s)

		// после Flush файл подключается как есть и ищет так же
		s.Flush(2)
		if Open(path, 1).Loaded() {
			t.Fatal("old stamp")
		}
		s = Open(path, 2)
		if !s.Loaded() || s.Count() != len(ref) {
			t.Fatal("not loaded", s.Count(), len(ref))
		}
		for _, q := range []string{"a", "b", "аб", "a @ru", "@a", "tag"} {
			if got, want := ids(s.Search(20, q)), ref.search(20, q); !slices.Equal(got, want) {
				t.Fatalf("%q:\n got %v\nwant %v", q, got, want)
			}
		}
		for _, x := range s.Search(1000, "a") {
			if x.Data != fmt.Sprint("data", x.ID) {
				t.Fatal("data", x)
			}
		}

		// битый файл не подключается
		os.WriteFile(path, []byte("plcidx01 garbage"), 0o644)
		if Open(path, 2).Loaded() {
			t.Fatal("garbage loaded")
		}
	})
}

func testRandom(t *testing.T, s *DB) naive {
	r := rand.New(rand.NewSource(1))
	ref := naive{}
	countries := []string{"@ru", "@us", "@fr", "@a", "ab", "tag"} // "@a" и "ab" бывают и словом, и опцией

	mutate := func() {
		id := r.Intn(3000)
		if r.Intn(5) == 0 {
			s.Delete(id)
			delete(ref, id)
			return
		}
		var text []string
		for n := 1 + r.Intn(3); n > 0; n-- {
			text = append(text, randWord(r, "abcаб", 6))
		}
		text = append(text, "@"+randWord(r, "ab", 4))
		var opts []string
		for n := r.Intn(3); n > 0; n-- {
			opts = append(opts, countries[r.Intn(len(countries))])
		}
		item := Item{ID: id, People: r.Intn(50), Data: fmt.Sprint("data", id)}
		if r.Intn(10) == 0 {
			item.Boost = int64(r.Intn(5))
		}
		s.Put(item, strings.Join(text, " "), opts)
		ref[id] = newSource(item, strings.Join(text, " "), opts)
	}

	check := func(step int) {
		for n := 0; n < 40; n++ {
			q := randWord(r, "abcаб", 3)
			switch r.Intn(6) {
			case 0:
				q += " " + countries[r.Intn(len(countries))]
			case 1:
				q += " @" + randWord(r, "ab", 2)
			case 2:
				q = countries[r.Intn(len(countries))]
			case 3:
				q = "@" + randWord(r, "ab", 2) + " " + countries[r.Intn(len(countries))]
			}
			limit := 1 + r.Intn(20)
			got, want := ids(s.Search(limit, q)), ref.search(limit, q)
			if !slices.Equal(got, want) {
				t.Fatalf("step %d, %q limit %d:\n got %v\nwant %v", step, q, limit, got, want)
			}
		}
		if s.Count() != len(ref) {
			t.Fatalf("step %d: count %d, want %d", step, s.Count(), len(ref))
		}
	}

	// пакетная загрузка
	s.Begin()
	for i := 0; i < 2000; i++ {
		mutate()
	}
	s.Commit()
	check(-1)

	// по одному: дельта и пересборки каждые compactAt изменений
	for step := 0; step < 1500; step++ {
		mutate()
		if step%7 == 0 {
			check(step)
		}
	}

	for _, x := range s.Search(1000, "a") {
		if x.Data != fmt.Sprint("data", x.ID) {
			t.Fatal("data", x)
		}
	}
	return ref
}

// поиск во время изменений, запускать с -race
func TestConcurrent(t *testing.T) {
	s := Open(filepath.Join(t.TempDir(), "search.idx"), 1)
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 2000; i++ {
				s.Add(i%700, i, 0, fmt.Sprintf("city%d new", i%50), []string{"@us"})
				if i%9 == 0 {
					s.Delete(i % 300)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 5000; i++ {
				s.Search(10, "ci @us")
			}
		}()
	}
	wg.Wait()
}

func TestNoAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("под -race sync.Pool намеренно теряет объекты")
	}
	s := New()
	for i := 0; i < 1000; i++ {
		s.Add(i, i, 0, fmt.Sprintf("new%d Нью%d", i, i), []string{"@us"})
	}
	var found int
	for _, q := range []string{"n", "new1", "new @us", "нью", "zzz"} {
		if n := testing.AllocsPerRun(100, func() { s.Each(10, q, func(x *Item) { found += x.ID }) }); n != 0 {
			t.Errorf("%q: %v allocs", q, n)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	s := New()
	b.ReportAllocs()
	s.Begin()
	for n := 0; n < b.N; n++ {
		s.Add(n, n, 10, fmt.Sprintf("c%d NewYork @12345", n), []string{"@ny"})
	}
	s.Commit()
}

func BenchmarkSearch(b *testing.B) {
	s := New()
	s.Begin()
	for n := 0; n < 200000; n++ {
		s.Add(n, n, 0, fmt.Sprintf("c%d NewYork%d @12345", n, n), []string{"@ny"})
	}
	s.Commit()
	var found int
	for _, q := range []string{"n", "new", "newyork1234", "new @ny"} {
		b.Run(q, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				s.Each(10, q, func(x *Item) { found += x.ID })
			}
		})
	}
}
