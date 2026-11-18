package place

import (
	"strings"

	"github.com/monopolly/file"
	"github.com/tidwall/gjson"
)

// get updates file json
func (a *Engine) Backup() (res []byte) {
	res = file.OpenE(a.updatefile)
	return
}

// json
func (a *Engine) Restore(changes []byte) {
	a.options.log.Println("restore changes...")

	// changes list
	list := make(map[int]P)

	//add new changes
	for _, b := range strings.Split(string(changes), "\n") {
		b = strings.TrimSpace(b)
		c := P([]byte(b))
		if !gjson.ValidBytes(c) {
			continue
		}
		id := c.ID()
		if id == 0 {
			continue
		}
		list[id] = c
	}

	// индекс и файл изменений обновляем один раз на всю пачку
	a.search.Begin()
	for _, p := range list {
		a.update(p)
		a.updates.set(p)
	}
	a.search.Commit()
	a.updates.save()
	a.options.log.Println("added", len(list))
}
