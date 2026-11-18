package model

import (
	"github.com/monopolly/numbers"
)

func (a *Place) BID() []byte {
	return numbers.IntBytes(a.ID)
}

// считаем search и options
func (a *P) Reindex(allowlang map[string]bool) {
	a.ReindexSearch(allowlang)
	a.ReindexOptions()
}

func (a *P) BID() []byte {
	return numbers.IntBytes(a.ID())
}
