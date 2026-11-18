package place

import (
	"log"

	"github.com/akrylysov/pogreb"
)

// store это база полных записей на диске: id > json
type store struct {
	db *pogreb.DB
}

func openStore(dir string, logger *log.Logger) (a *store, err error) {
	a = new(store)
	pogreb.SetLogger(logger)
	a.db, err = pogreb.Open(dir, nil)
	return
}

func (a *store) Get(k []byte) (v []byte) {
	v, _ = a.db.Get(k)
	return
}

func (a *store) Put(k, v []byte) {
	a.db.Put(k, v)
}

func (a *store) Delete(k []byte) {
	a.db.Delete(k)
}

func (a *store) Count() int {
	return int(a.db.Count())
}

func (a *store) Close() {
	a.db.Close()
}

// Iterate проходит по всей базе
func (a *store) Iterate(f func(k, v []byte)) {
	it := a.db.Items()
	for {
		k, v, err := it.Next()
		if err != nil {
			return
		}
		f(k, v)
	}
}
