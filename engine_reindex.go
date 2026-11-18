package place

import (
	"bytes"
	"time"
)

// reindex all db
func (a *Engine) Reindex() {
	a.options.log.Println("reindex...")
	t := time.Now()
	allow := a.options.langmap
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
			p.Set("updates", true)
		}

		p.Reindex(allow)
		if !bytes.Equal(p, v) {
			// базу трогаем, только если запись изменилась
			a.addToDB(p)
		}
		a.addToSearch(p)
	})
	a.search.Commit()
	a.options.log.Println("...", time.Since(t))
	a.options.log.Println("languages:", a.options.langs)
}
