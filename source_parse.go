package place

import (
	"fmt"
	"time"
)

// только добавляет в базу все из geonames cities500.txt
//
// Память важнее скорости, поэтому файл городов читается дважды: сначала только
// id, потом, когда названия собраны, города по одному уходят в базу.
func (a *Engine) parse() (err error) {
	t1 := time.Now()
	a.options.log.Println("parse city...")
	index := make(map[int]int32) // id > номер города
	err = a.lines(a.options.source, func(line []byte) {
		if p := a.parseCitiesLine(string(line)); p != nil {
			if _, ok := index[p.ID]; !ok {
				index[p.ID] = int32(len(index))
			}
		}
	})
	if err != nil {
		return
	}
	if len(index) == 0 {
		return fmt.Errorf("%s: no cities", a.options.source.name())
	}
	a.options.log.Println("done parse city:", len(index), time.Since(t1))

	// названия нужны только для этих городов
	t2 := time.Now()
	a.options.log.Println("parse names...")
	names, err := a.parseNames(alternateNamesV2, index)
	if err != nil {
		// без переводов база все равно рабочая
		a.options.log.Println("parse names error...", err)
	}
	a.options.log.Println("done parse names:", time.Since(t2))

	allow := a.options.langmap
	return a.lines(a.options.source, func(line []byte) {
		c := a.parseCitiesLine(string(line))
		if c == nil {
			return
		}
		if i := index[c.ID]; int(i) < len(names) {
			c.Names = decodeNames(names[i])
			names[i] = nil
		}
		// сразу с поисковыми полями, чтобы Reindex не переписывал запись
		p := P(c.ToJson())
		p.Reindex(allow)
		a.db.Put(p.BID(), p)
	})
}
