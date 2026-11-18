package place

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// Бенчмарки на реальных данных geonames.
//
//	scripts/download.sh            # один раз
//	go test -run TestBenchInit -v  # время старта и память
//	go test -run xxx -bench . -benchmem
//
// PLACE_SOURCE=cities500 переключает набор городов (по умолчанию cities15000).

var benchQueries = []string{"n", "s", "new", "san", "new york", "мос", "лон", "ве @ru", "pa @fr", "s @us", "saint petersburg", "zzzz"}

var benchFields = []string{FieldFlag, FieldName, FieldCountry, FieldPeople, "location"}

func benchOptions(t testing.TB) *options {
	data := os.Getenv("PLACE_DATA")
	if data == "" {
		data = "testdata"
	}
	src := os.Getenv("PLACE_SOURCE")
	if src == "" {
		src = "cities15000"
	}
	if _, err := os.Stat(filepath.Join(data, dirSource, src+".zip")); err != nil {
		t.Skip("нет данных geonames, запусти scripts/download.sh:", err)
	}
	// у каждого набора своя база, исходники общие
	volume := filepath.Join(data, "bench-"+src)
	os.MkdirAll(volume, 0o755)
	if _, err := os.Lstat(filepath.Join(volume, dirSource)); err != nil {
		abs, _ := filepath.Abs(filepath.Join(data, dirSource))
		if err := os.Symlink(abs, filepath.Join(volume, dirSource)); err != nil {
			t.Fatal(err)
		}
	}
	o := NewOptions().Langs("en", "ru", "it", "es", "fr").Volume(volume)
	switch src {
	case "cities500":
		o.City500()
	case "cities1000":
		o.City1000()
	case "cities5000":
		o.City5000()
	default:
		o.City15000()
	}
	if os.Getenv("PLACE_VERBOSE") == "" {
		o.Log(log.New(io.Discard, "", 0))
	}
	return o
}

func heapMB() float64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / (1 << 20)
}

// сколько кучи процесс брал у системы на пике (без mmap файлов)
func heapPeakMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapSys) / (1 << 20)
}

// время старта и память, которую держит движок
func TestBenchInit(t *testing.T) {
	o := benchOptions(t)
	before := heapMB()
	t0 := time.Now()
	e := New(o)
	defer e.Close()
	fmt.Printf("init: %v, places: %d, heap: %.1f MB, heap peak: %.0f MB\n", time.Since(t0).Round(time.Millisecond), e.Count(), heapMB()-before, heapPeakMB())
	for _, q := range benchQueries {
		t1 := time.Now()
		res := e.Search(5, q, "", benchFields...)
		fmt.Printf("%-18q %10v %s\n", q, time.Since(t1), res)
	}
	runtime.KeepAlive(e)
}

var (
	benchOnce   sync.Once
	benchEngine *Engine
)

// базу надо закрыть, иначе следующий запуск начнется с восстановления pogreb
func TestMain(m *testing.M) {
	code := m.Run()
	if benchEngine != nil {
		benchEngine.Close()
	}
	os.Exit(code)
}

func benchNew(b *testing.B) *Engine {
	o := benchOptions(b)
	benchOnce.Do(func() { benchEngine = New(o) })
	return benchEngine
}

// только индекс: текст запроса > id
func BenchmarkIndex(b *testing.B) {
	e := benchNew(b)
	for _, q := range benchQueries {
		b.Run(q, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				e.search.Search(10, q)
			}
		})
	}
}

// весь путь запроса: индекс + сборка json
func BenchmarkSearch(b *testing.B) {
	e := benchNew(b)
	for _, q := range benchQueries {
		b.Run(q, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				e.Search(10, q, "", benchFields...)
			}
		})
	}
}

func BenchmarkSearchParallel(b *testing.B) {
	e := benchNew(b)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var n int
		for pb.Next() {
			e.Search(10, benchQueries[n%len(benchQueries)], "", benchFields...)
			n++
		}
	})
}
