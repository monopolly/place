package place

import (
	"bytes"
	_ "embed"
	"strings"
	"sync"

	"github.com/monopolly/file"
	"github.com/tidwall/gjson"
)

func newDeleted(filename string) (a *deleted) {
	a = new(deleted)
	a.file = filename
	a.list = make(map[int]P)
	return
}

type deleted struct {
	file string
	list map[int]P
	sync.RWMutex
}

// save
func (a *deleted) save() {
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
func (a *deleted) add(p P) {
	a.set(p)
	a.save()
}

// update without save
func (a *deleted) set(p P) {
	a.Lock()
	a.list[p.ID()] = p
	a.Unlock()
}

func (a *deleted) get(id int) P {
	a.RLock()
	defer a.RUnlock()
	return a.list[id]
}

func (a *deleted) has(id int) bool {
	return a.get(id) != nil
}

// delete
func (a *deleted) delete(id int) {
	a.Lock()
	delete(a.list, id)
	a.Unlock()
	a.save()
}

func (a *deleted) exist() bool {
	return file.Exists(a.file)
}

// load
func (a *deleted) init() {

	var errors int
	// open changes
	if !a.exist() {
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
		list[id] = p
	})

}

// load
func (a *deleted) initDefault(b []byte) {

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
