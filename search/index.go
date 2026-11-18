package search

import (
	"cmp"
	"math"
	"math/bits"
	"slices"
	"strings"
)

// Item это результат поиска
type Item struct {
	ID     int
	People int
	Boost  int64
	Data   string // произвольные данные, переданные в Put
}

// порядок выдачи: сначала boost (новые выше), потом население
func compare(a, b *Item) int {
	if a.Boost != b.Boost {
		return cmp.Compare(b.Boost, a.Boost)
	}
	if a.People != b.People {
		return cmp.Compare(b.People, a.People)
	}
	return cmp.Compare(a.ID, b.ID)
}

// source это документ до индексации. Слова и опции лежат одной строкой:
// при загрузке таких документов сотни тысяч, и память здесь важнее удобства.
type source struct {
	Item
	text  string // слова через пробел (ищутся по префиксу), дальше опции через 0 (целиком)
	split uint32 // где кончаются слова
}

func (a *source) terms() string { return a.text[:a.split] }
func (a *source) opts() string  { return a.text[a.split:] }

func (a *source) match(words [][]byte) bool {
next:
	for _, w := range words {
		for t, rest := next(a.terms(), ' '); t != ""; t, rest = next(rest, ' ') {
			if hasPrefix(t, w) {
				continue next
			}
		}
		for o, rest := next(a.opts(), 0); o != ""; o, rest = next(rest, 0) {
			if o == string(w) {
				continue next
			}
		}
		return false
	}
	return true
}

// dict это отсортированный словарь со списками документов.
// Номер документа это его место в общей выдаче (ранг), поэтому
// список любого слова уже отсортирован так, как нужно отдавать.
type dict struct {
	blob  string   // все слова подряд
	off   []uint32 // границы слов в blob, len = n+1
	start []uint32 // границы списков в docs, len = n+1
	docs  []uint32 // ранги документов, внутри слова по возрастанию
}

func (a *dict) len() int { return len(a.start) - 1 }

func (a *dict) term(i uint32) string { return a.blob[a.off[i]:a.off[i+1]] }

// find ищет слово целиком
func (a *dict) find(w []byte) (id uint32, ok bool) {
	lo, hi := uint32(0), uint32(a.len())
	for lo < hi {
		m := (lo + hi) / 2
		if a.term(m) < string(w) {
			lo = m + 1
		} else {
			hi = m
		}
	}
	return lo, int(lo) < a.len() && a.term(lo) == string(w)
}

// index это неизменяемый снимок индекса. Все данные лежат в плоских массивах
// без указателей, поэтому индекс целиком пишется в файл и читается через mmap:
// в куче Go он тогда не занимает ничего, а память под него выделяет и при
// нехватке забирает обратно операционная система.
type index struct {
	// документы по рангу
	ids     []int64
	boost   []int64
	people  []uint32
	dataOff []uint32 // границы Data в data, len = n+1
	data    string

	terms dict
	opts  dict

	// слова и опции каждого документа, чтобы проверять фильтры без строк
	docTermStart, docTerms []uint32
	docOptStart, docOpts   []uint32

	first []uint32 // первое слово для каждой пары начальных байт, len = 65537
	rmq   rmq      // минимум ранга на отрезке terms.docs

	stamp uint64 // метка состояния, с которым записан файл
	unmap func() // освобождает mmap, nil если индекс в куче
}

func (a *index) len() int { return len(a.ids) }

func (a *index) item(doc uint32) Item {
	return Item{
		ID:     int(a.ids[doc]),
		People: int(a.people[doc]),
		Boost:  a.boost[doc],
		Data:   a.data[a.dataOff[doc]:a.dataOff[doc+1]],
	}
}

// build собирает индекс. Сами документы не меняет: их может читать поиск.
func build(src []*source) (a *index) {
	a = new(index)
	slices.SortFunc(src, func(x, y *source) int { return compare(&x.Item, &y.Item) })
	n := len(src)
	var size int
	for i := range src {
		size += len(src[i].Data)
	}
	a.ids = make([]int64, n)
	a.boost = make([]int64, n)
	a.people = make([]uint32, n)
	a.dataOff = make([]uint32, n+1)
	var data strings.Builder
	data.Grow(size)
	for i := range src {
		x := &src[i].Item
		a.ids[i] = int64(x.ID)
		a.boost[i] = x.Boost
		a.people[i] = uint32(min(max(x.People, 0), math.MaxUint32))
		data.WriteString(x.Data)
		a.dataOff[i+1] = uint32(data.Len())
	}
	a.data = data.String()
	a.terms, a.docTermStart, a.docTerms = buildDict(src, false)
	a.opts, a.docOptStart, a.docOpts = buildDict(src, true)

	a.first = make([]uint32, 1<<16+1)
	var k int
	for i := 0; i < a.terms.len(); i++ {
		for key := key2(a.terms.term(uint32(i))); k <= key; k++ {
			a.first[k] = uint32(i)
		}
	}
	for ; k < len(a.first); k++ {
		a.first[k] = uint32(a.terms.len())
	}

	a.rmq.build(a.terms.docs)
	return
}

func key2(s string) (k int) {
	if len(s) > 0 {
		k = int(s[0]) << 8
	}
	if len(s) > 1 {
		k |= int(s[1])
	}
	return
}

func key8(s string) (k uint64) {
	for i := 0; i < 8; i++ {
		k <<= 8
		if i < len(s) {
			k |= uint64(s[i])
		}
	}
	return
}

// buildDict собирает словарь слов (или опций) всех документов
func buildDict(src []*source, opts bool) (d dict, docStart, docIDs []uint32) {
	// слово это ссылка на кусок text своего документа
	type entry struct {
		key uint64 // первые 8 байт слова: почти все сравнения обходятся без строк
		doc uint32
		at  uint32 // позиция в text << 12 | длина
	}
	str := func(e entry) string {
		pos, n := e.at>>12, e.at&maxWord
		return src[e.doc].text[pos : pos+n]
	}
	each := func(fn func(doc, pos int, w string)) {
		for i, s := range src {
			list, sep, base := s.terms(), byte(' '), 0
			if opts {
				list, sep, base = s.opts(), 0, int(s.split)
			}
			for pos := 0; pos < len(list); {
				w, _ := next(list[pos:], sep)
				fn(i, base+pos, w)
				pos += len(w) + 1
			}
		}
	}

	var total, size int
	each(func(_, _ int, w string) {
		total++
		size += len(w)
	})
	list := make([]entry, 0, total)
	each(func(doc, pos int, w string) {
		list = append(list, entry{key8(w), uint32(doc), uint32(pos)<<12 | uint32(len(w))})
	})
	slices.SortFunc(list, func(x, y entry) int {
		if x.key != y.key {
			return cmp.Compare(x.key, y.key)
		}
		if c := strings.Compare(str(x), str(y)); c != 0 {
			return c
		}
		return cmp.Compare(x.doc, y.doc)
	})

	var blob strings.Builder
	blob.Grow(size)
	d.docs = make([]uint32, 0, total)
	d.off = append(d.off, 0)
	docStart = make([]uint32, len(src)+1)
	var prev string
	for i, e := range list {
		w := str(e)
		switch {
		case i == 0 || w != prev:
			d.start = append(d.start, uint32(len(d.docs)))
			blob.WriteString(w)
			d.off = append(d.off, uint32(blob.Len()))
			prev = w
		case e.doc == list[i-1].doc:
			continue
		}
		d.docs = append(d.docs, e.doc)
		docStart[e.doc+1]++
	}
	d.start = append(d.start, uint32(len(d.docs)))
	d.blob = blob.String()
	d.off = slices.Clip(d.off)
	d.start = slices.Clip(d.start)
	d.docs = slices.Clip(d.docs)

	// документ > номера его слов
	for i := 1; i < len(docStart); i++ {
		docStart[i] += docStart[i-1]
	}
	docIDs = make([]uint32, len(d.docs))
	pos := slices.Clone(docStart[:len(src)])
	for t := 0; t < d.len(); t++ {
		for _, doc := range d.docs[d.start[t]:d.start[t+1]] {
			docIDs[pos[doc]] = uint32(t)
			pos[doc]++
		}
	}
	return
}

// source восстанавливает документ из индекса для пересборки
func (a *index) source(doc int) *source {
	var b []byte
	for _, t := range a.docTerms[a.docTermStart[doc]:a.docTermStart[doc+1]] {
		if len(b) > 0 {
			b = append(b, ' ')
		}
		b = append(b, a.terms.term(t)...)
	}
	split := len(b)
	for _, o := range a.docOpts[a.docOptStart[doc]:a.docOptStart[doc+1]] {
		if len(b) > split {
			b = append(b, 0)
		}
		b = append(b, a.opts.term(o)...)
	}
	return &source{Item: a.item(uint32(doc)), text: string(b), split: uint32(split)}
}

// prefix возвращает диапазон слов, которые начинаются с w
func (a *index) prefix(w []byte) (lo, hi uint32) {
	switch len(w) {
	case 0:
		return 0, 0
	case 1:
		k := int(w[0]) << 8
		return a.first[k], a.first[k+256]
	}
	k := int(w[0])<<8 | int(w[1])
	lo, hi = a.first[k], a.first[k+1]
	if len(w) == 2 {
		return
	}
	end := hi
	for lo < hi {
		m := (lo + hi) / 2
		if a.terms.term(m) < string(w) {
			lo = m + 1
		} else {
			hi = m
		}
	}
	hi = end
	from := lo
	for lo < hi {
		m := (lo + hi) / 2
		if hasPrefix(a.terms.term(m), w) {
			lo = m + 1
		} else {
			hi = m
		}
	}
	return from, lo
}

// cond это одно слово запроса: подходят документы, у которых есть слово
// с таким префиксом или точно такая опция
type cond struct {
	tlo, thi uint32   // номера слов с префиксом
	elo, ehi uint32   // их списки в terms.docs
	opt      []uint32 // список документов с опцией
	optID    uint32
	hasOpt   bool
}

func (c *cond) size() int { return int(c.ehi-c.elo) + len(c.opt) }

func (a *index) has(doc uint32, c *cond) bool {
	for _, t := range a.docTerms[a.docTermStart[doc]:a.docTermStart[doc+1]] {
		if t >= c.thi {
			break
		}
		if t >= c.tlo {
			return true
		}
	}
	if c.hasOpt {
		for _, o := range a.docOpts[a.docOptStart[doc]:a.docOptStart[doc+1]] {
			if o == c.optID {
				return true
			}
		}
	}
	return false
}

// search дописывает в dst до limit лучших документов, где есть все слова
func (a *index) search(dst []Item, limit int, words [][]byte, dead map[int]struct{}, sc *scratch) []Item {
	if a.len() == 0 || len(words) == 0 {
		return dst
	}
	var cb [maxWords]cond
	conds := cb[:len(words)]
	var main int // самое редкое слово, идем по нему
	for i, w := range words {
		c := &conds[i]
		c.tlo, c.thi = a.prefix(w)
		c.elo, c.ehi = a.terms.start[c.tlo], a.terms.start[c.thi]
		if c.optID, c.hasOpt = a.opts.find(w); c.hasOpt {
			c.opt = a.opts.docs[a.opts.start[c.optID]:a.opts.start[c.optID+1]]
		}
		if c.size() == 0 {
			return dst
		}
		if c.size() < conds[main].size() {
			main = i
		}
	}

	// документы главного слова в порядке ранга: слияние отрезка префикса
	// и списка опции, оба отдают ранги по возрастанию
	c := &conds[main]
	it := topIter{rmq: &a.rmq, heap: sc.heap[:0]}
	defer func() { sc.heap = it.heap }()
	it.push(c.elo, c.ehi)
	rv, rok := it.next()
	opt := c.opt
	last := uint32(math.MaxUint32)
	end := len(dst) + limit
next:
	for len(dst) < end {
		var doc uint32
		switch {
		case rok && (len(opt) == 0 || rv <= opt[0]):
			doc = rv
			rv, rok = it.next()
		case len(opt) > 0:
			doc = opt[0]
			opt = opt[1:]
		default:
			break next
		}
		if doc == last {
			continue
		}
		last = doc
		for i := range conds {
			if i != main && !a.has(doc, &conds[i]) {
				continue next
			}
		}
		if dead != nil {
			if _, ok := dead[int(a.ids[doc])]; ok {
				continue
			}
		}
		dst = append(dst, a.item(doc))
	}
	return dst
}

// rmq отвечает, где на отрезке массива стоит минимум: блоки по 16 значений
// и разреженная таблица поверх минимумов блоков
type rmq struct {
	a    []uint32
	bmin []uint32 // минимум каждого полного блока
	sp   []uint32 // уровни таблицы подряд, начиная с первого
	lv   []uint32 // начало уровня j в sp: блок с минимумом среди [i, i+2^j)
}

const (
	blockShift = 4
	blockSize  = 1 << blockShift
)

func (r *rmq) build(a []uint32) {
	nb := len(a) >> blockShift
	bmin := make([]uint32, nb)
	for b := range bmin {
		_, bmin[b] = scan(a, uint32(b)<<blockShift, uint32(b+1)<<blockShift)
	}
	var size int
	for j := 1; 1<<j <= nb; j++ {
		size += nb - 1<<j + 1
	}
	sp := make([]uint32, 0, size)
	var prev []uint32 // nil для нулевого уровня: там блок i это i
	for j := 1; 1<<j <= nb; j++ {
		from := len(sp)
		for i := 0; i <= nb-1<<j; i++ {
			x, y := uint32(i), uint32(i+1<<(j-1))
			if prev != nil {
				x, y = prev[x], prev[y]
			}
			if bmin[y] < bmin[x] {
				x = y
			}
			sp = append(sp, x)
		}
		prev = sp[from:]
	}
	r.init(a, bmin, sp)
}

func (r *rmq) init(a, bmin, sp []uint32) {
	r.a, r.bmin, r.sp = a, bmin, sp
	r.lv = []uint32{0, 0} // нулевой уровень не хранится
	nb := len(bmin)
	var from int
	for j := 1; 1<<j <= nb; j++ {
		from += nb - 1<<j + 1
		r.lv = append(r.lv, uint32(from))
	}
}

func scan(a []uint32, lo, hi uint32) (pos, val uint32) {
	val = math.MaxUint32
	for i, v := range a[lo:hi] {
		if v < val {
			pos, val = lo+uint32(i), v
		}
	}
	return
}

// argmin возвращает позицию и значение минимума на [lo, hi), lo < hi
func (r *rmq) argmin(lo, hi uint32) (pos, val uint32) {
	if hi-lo <= 2*blockSize {
		return scan(r.a, lo, hi)
	}
	bl := (lo + blockSize - 1) >> blockShift
	bh := hi >> blockShift
	pos, val = scan(r.a, lo, bl<<blockShift)
	if p, v := scan(r.a, bh<<blockShift, hi); v < val {
		pos, val = p, v
	}
	j := bits.Len32(bh-bl) - 1
	x, y := bl, bh-1
	if j > 0 {
		level := r.sp[r.lv[j]:]
		x, y = level[bl], level[bh-1<<j]
	}
	if r.bmin[y] < r.bmin[x] {
		x = y
	}
	if r.bmin[x] < val {
		pos, val = scan(r.a, x<<blockShift, (x+1)<<blockShift)
	}
	return
}

// topIter отдает значения отрезка по возрастанию: достает минимум и делит
// отрезок на части слева и справа от него
type topIter struct {
	rmq  *rmq
	heap []span
}

type span struct{ val, pos, lo, hi uint32 }

func (a *topIter) push(lo, hi uint32) {
	if lo >= hi {
		return
	}
	pos, val := a.rmq.argmin(lo, hi)
	h := append(a.heap, span{val, pos, lo, hi})
	for i := len(h) - 1; i > 0; {
		p := (i - 1) / 2
		if h[p].val <= h[i].val {
			break
		}
		h[p], h[i] = h[i], h[p]
		i = p
	}
	a.heap = h
}

func (a *topIter) next() (val uint32, ok bool) {
	h := a.heap
	if len(h) == 0 {
		return
	}
	top := h[0]
	n := len(h) - 1
	h[0] = h[n]
	h = h[:n]
	for i := 0; ; {
		l := 2*i + 1
		if l >= n {
			break
		}
		if r := l + 1; r < n && h[r].val < h[l].val {
			l = r
		}
		if h[i].val <= h[l].val {
			break
		}
		h[i], h[l] = h[l], h[i]
		i = l
	}
	a.heap = h
	a.push(top.lo, top.pos)
	a.push(top.pos+1, top.hi)
	return top.val, true
}
