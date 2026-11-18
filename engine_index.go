package place

import (
	"time"
)

// index db
func (a *Engine) index() {
	a.options.log.Println("index start...")
	t := time.Now()
	a.search.Begin()
	a.db.Iterate(func(k, v []byte) {
		if string(k) == hashKey {
			return
		}
		p := P(v)
		id := p.ID()
		if a.skipped(id) {
			a.options.log.Println("skip", p.Flag(), id, p.Name(), p.Country())
			return
		}
		if a.deleted.has(id) {
			a.options.log.Println("deleted", p.Flag(), id, p.Name(), p.Country())
			return
		}

		if u := a.updates.get(id); u != nil {
			p = u
		}

		a.addToSearch(p)
	})
	a.search.Commit()

	a.options.log.Println("index done", time.Since(t))
}
