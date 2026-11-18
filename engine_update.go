package place

import (
	"time"

	"github.com/monopolly/cast"
	"github.com/monopolly/errors"
	"github.com/monopolly/numbers"
)

func (a *Engine) Update(p P) {
	a.update(p)
	a.updates.add(p)
}

func (a *Engine) update(p P) {
	p.UP(time.Now().Unix())
	p.Set("updates", true)
	a.addToDB(p)
	a.addToSearch(p)
}
func (a *Engine) UpdateBytes(p []byte) {
	a.Update(P(p))
}

func (a *Engine) addToDB(p P) {
	a.db.Put(p.BID(), p)
}

func (a *Engine) addToSearch(p P) {
	a.search.Put(a.parsePlace(p))
}

func (a *Engine) Boost(id int) (err errors.E) {
	var b P
	b = a.db.Get(numbers.IntBytes(id))
	if b == nil {
		return errors.NotFound()
	}
	b.Boost(time.Now().Unix())
	a.Update(b)
	return
}

func (a *Engine) RemoveBoost(id int) {
	var b P
	b = a.db.Get(numbers.IntBytes(id))
	if b == nil {
		return
	}
	b.Boost(0)
	a.Update(b)
}

// key
func (a *Engine) UpdateKey(id int, key string, value any) {
	var b P
	b = a.db.Get(numbers.IntBytes(id))
	if b == nil {
		return
	}

	switch key {
	case FieldName:
		b.Name(cast.String(value))
	case FieldLat:
		b.Lat(cast.Float(value))
	case FieldLong:
		b.Long(cast.Float(value))
	case FieldPeople:
		b.People(cast.Int(value))
	case FieldState:
		b.State(cast.String(value))
	case FieldAlias:
		b.Alias(cast.String(value))
	case FieldTags:
		b.Tags(cast.String(value))
	case FieldDel:
		b.Del(cast.Bool(value))
	default:
		return
	}

	b.Reindex(a.options.langmap)
	a.Update(b)
}

// lang
func (a *Engine) UpdateNames(id int, lang string, v string) {
	var b P
	b = a.db.Get(numbers.IntBytes(id))
	if b == nil {
		return
	}

	names := b.Names()
	names[lang] = v
	b.Names(names)

	b.Reindex(a.options.langmap)
	a.Update(b)
}
