package place

import (
	"bytes"
	_ "embed"
	"strings"
	"sync"

	"github.com/monopolly/file"
	"github.com/tidwall/gjson"
)

//go:embed embed/changes.json
var defaultUpdates []byte

func newUpdates(filename string) (a *updates) {
	a = new(updates)
	a.file = filename
	a.list = make(map[int]P)
	return
}

type updates struct {
	file string
	list map[int]P
	sync.RWMutex
}

// save
func (a *updates) save() {
	var b bytes.Buffer
	a.RLock()
	for _, p := range a.list {
		b.Write(p)
		b.WriteString("\n")
	}
	a.RUnlock()
	file.Save(a.file, b.Bytes())
}

// update
func (a *updates) add(p P) {
	a.set(p)
	a.save()
}

// update without save
func (a *updates) set(p P) {
	a.Lock()
	a.list[p.ID()] = p
	a.Unlock()
}

func (a *updates) get(id int) P {
	a.RLock()
	defer a.RUnlock()
	return a.list[id]
}

// delete
func (a *updates) delete(id int) {
	a.Lock()
	delete(a.list, id)
	a.Unlock()
	a.save()
}

func (a *updates) exist() bool {
	return file.Exists(a.file)
}

// load
func (a *updates) init() {

	var errors int
	// open changes
	if !a.exist() {
		a.initDefault(defaultUpdates)
		return
	}

	list := make(map[int]P)
	defer func() {
		a.Lock()
		a.list = list
		a.Unlock()
	}()
	file.Play(a.file, func(b string) {
		b = strings.TrimSpace(b)
		p := P([]byte(b))
		if !gjson.ValidBytes(p) {
			errors++
			return
		}

		id := p.ID()
		if id == 0 {
			return
		}
		if p.Country() == "gb" {
			p.Country("uk")
		}
		list[id] = p
	})

}

// load
func (a *updates) initDefault(b []byte) {

	a.list = make(map[int]P)
	for _, b := range strings.Split(string(b), "\n") {
		b = strings.TrimSpace(b)
		p := P([]byte(b))
		if !gjson.ValidBytes(p) {
			continue
		}
		id := p.ID()
		if id == 0 {
			continue
		}
		a.list[id] = p
	}

}
