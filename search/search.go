// Package search это поиск по префиксу для автодополнения.
//
// Индекс неизменяемый: отсортированный словарь слов со списками документов,
// где номер документа это его место в выдаче. Поиск — бинарный поиск диапазона
// слов и выбор лучших документов на отрезке, без блокировок и аллокаций.
// Изменения копятся в небольшой дельте поверх индекса и вливаются в него
// пересборкой, когда их становится много.
//
// Индекс, открытый через Open, живет в файле и подключается через mmap,
// так что в куче остается только дельта.
package search

import (
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
)

// сколько изменений держим в дельте до пересборки индекса
const compactAt = 256

// New создает индекс в памяти
func New() (a *DB) {
	a = new(DB)
	a.pend = make(map[int]*source)
	a.view.Store(&view{ix: build(nil)})
	return
}

// Open создает индекс в файле path. Если файл есть и записан с той же меткой
// stamp, он подключается как есть и Loaded возвращает true. Иначе индекс пуст
// и его надо наполнить. Метка это отпечаток состояния данных, из которых
// индекс собран: по ней видно, что файл устарел.
func Open(path string, stamp uint64) (a *DB) {
	a = New()
	a.path = path
	a.stamp = stamp
	if ix, err := load(path); err == nil && ix.stamp == stamp {
		a.view.Store(&view{ix: ix})
		a.loaded = true
	}
	return
}

type DB struct {
	view atomic.Pointer[view] // то, что видит поиск

	mu     sync.Mutex
	pend   map[int]*source // изменения поверх индекса, nil значит удален
	batch  int
	path   string // файл индекса, пусто если индекс в памяти
	stamp  uint64
	loaded bool
	gc     int // настройка сборщика мусора до Begin
}

type view struct {
	ix    *index
	delta []*source        // добавленные и измененные после сборки индекса
	dead  map[int]struct{} // id, которые в индексе больше не действуют
}

// Loaded сообщает, что Open подключил готовый файл
func (a *DB) Loaded() bool { return a.loaded }

// Add добавляет или заменяет документ. text это слова, которые ищутся
// по префиксу, options ищутся только целиком (@ru, 90210).
func (a *DB) Add(id int, population int, boost int64, text string, options []string) {
	a.Put(Item{ID: id, People: population, Boost: boost}, text, options)
}

// Put это Add с произвольными данными, которые вернутся в результате поиска.
// Строку Data индекс оставляет себе как есть.
func (a *DB) Put(item Item, text string, options []string) {
	s := newSource(item, text, options)
	a.mu.Lock()
	a.pend[item.ID] = s
	a.publish(false)
	a.mu.Unlock()
}

func (a *DB) Delete(id int) {
	a.mu.Lock()
	a.pend[id] = nil
	a.publish(false)
	a.mu.Unlock()
}

// Begin откладывает публикацию изменений до Commit: массовая загрузка
// собирает индекс один раз. Поиск до Commit видит прежнее состояние.
//
// На время загрузки сборщик мусора работает чаще: загрузка большая, и пик
// памяти процесса важнее ее скорости.
func (a *DB) Begin() {
	a.mu.Lock()
	if a.batch++; a.batch == 1 {
		a.gc = debug.SetGCPercent(batchGC)
	}
	a.mu.Unlock()
}

func (a *DB) Commit() {
	a.mu.Lock()
	a.batch--
	a.publish(false)
	if a.batch == 0 {
		debug.SetGCPercent(a.gc)
	}
	a.mu.Unlock()
}

const batchGC = 20

// Flush вливает дельту в индекс и записывает файл с меткой stamp, чтобы
// следующий Open подключил его без пересборки. Вызывать перед выходом.
func (a *DB) Flush(stamp uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stamp = stamp
	if a.path == "" {
		return
	}
	ix := a.view.Load().ix
	switch {
	case len(a.pend) > 0 || ix.unmap == nil:
		a.publish(true)
	case ix.stamp != stamp:
		if writeStamp(a.path, stamp) == nil {
			ix.stamp = stamp
		}
	}
}

// Reset очищает индекс
func (a *DB) Reset() {
	a.mu.Lock()
	a.pend = make(map[int]*source)
	a.view.Store(&view{ix: build(nil)})
	a.loaded = false
	a.mu.Unlock()
}

// Count это количество документов
func (a *DB) Count() (n int) {
	v := a.view.Load()
	n = v.ix.len() + len(v.delta)
	if len(v.dead) > 0 {
		for _, id := range v.ix.ids {
			if _, ok := v.dead[int(id)]; ok {
				n--
			}
		}
	}
	return
}

// publish показывает поиску накопленные изменения: дельтой, пока их мало,
// иначе пересборкой индекса
func (a *DB) publish(compact bool) {
	if a.batch > 0 {
		return
	}
	old := a.view.Load()
	v := &view{ix: old.ix}
	if compact || len(a.pend) > compactAt {
		src := make([]*source, 0, old.ix.len()+len(a.pend))
		for i := 0; i < old.ix.len(); i++ {
			if _, ok := a.pend[int(old.ix.ids[i])]; !ok {
				src = append(src, old.ix.source(i))
			}
		}
		for _, s := range a.pend {
			if s != nil {
				src = append(src, s)
			}
		}
		a.pend = make(map[int]*source)
		v.ix = build(src)
		runtime.KeepAlive(old) // строки src смотрели в старый индекс
		if a.path != "" {
			// собранное в куче меняем на файл; не вышло — остаемся в куче
			if err := v.ix.write(a.path, a.stamp); err == nil {
				if ix, err := load(a.path); err == nil {
					v.ix = ix
				}
			}
		}
		a.view.Store(v)
		src = nil
		debug.FreeOSMemory() // сборка мусорит, память нужна системе сразу
		return
	}
	if len(a.pend) > 0 {
		v.dead = make(map[int]struct{}, len(a.pend))
		for id, s := range a.pend {
			v.dead[id] = struct{}{}
			if s != nil {
				v.delta = append(v.delta, s)
			}
		}
	}
	a.view.Store(v)
}

// Search возвращает до limit лучших документов: сначала с boost, потом
// по населению. Слова с '@' это фильтры: "ве @ru".
func (a *DB) Search(limit int, text string) (res []Item) {
	a.Each(limit, text, func(x *Item) {
		x.Data = strings.Clone(x.Data)
		res = append(res, *x)
	})
	return
}

// Each это Search без аллокаций: вызывает fn для каждого результата по порядку.
// Item и его Data действуют только внутри fn.
func (a *DB) Each(limit int, text string, fn func(x *Item)) {
	if limit <= 0 {
		return
	}
	sc := pool.Get().(*scratch)
	defer pool.Put(sc)
	words := sc.parseQuery(text)
	if len(words) == 0 {
		return
	}

	v := a.view.Load()
	list := v.ix.search(sc.items[:0], limit, words, v.dead, sc)

	// документы дельты вставляем на свои места, не выходя за limit
	for _, s := range v.delta {
		if !s.match(words) {
			continue
		}
		i := len(list)
		for i > 0 && compare(&s.Item, &list[i-1]) < 0 {
			i--
		}
		if i == limit {
			continue
		}
		if len(list) < limit {
			list = append(list, Item{})
		}
		copy(list[i+1:], list[i:])
		list[i] = s.Item
	}

	for i := range list {
		fn(&list[i])
	}
	clear(list) // не держим строки из индекса в пуле
	sc.items = list[:0]
	runtime.KeepAlive(v) // Data смотрит в память индекса
}
