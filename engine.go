package place

import (
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/monopolly/file"
	"github.com/monopolly/jsons"
	"github.com/monopolly/numbers"
	"github.com/monopolly/place/model"
	"github.com/monopolly/place/search"
)

type P = model.P
type Place = model.Place

const (
	// typeCity  = "city"
	// typePlace = "place"

	dirDB      = "index"
	dirSource  = "source"
	dirUpdates = "updates"
)

var (
	volume = "db"
)

func New(options *options) (a *Engine) {
	a = new(Engine)
	a.options = options
	a.options.prepare()
	volume = a.options.volume
	a.zipcodes = make(map[int]*P)

	// dir
	a.options.log.Println("create dirs...")
	file.Dir(volume, dirSource, dirDB, dirUpdates)

	// db
	a.options.log.Println("init db...")
	var err error
	if a.db, err = openStore(path.Join(volume, dirDB), a.options.log); err != nil {
		panic(fmt.Errorf("place: open db: %w", err))
	}

	// updates
	a.options.log.Println("init updates...")
	a.updatefile = filepath.Join(volume, dirUpdates, "updates.json")
	a.namesfile = filepath.Join(volume, dirUpdates, "names.json")
	a.deletefile = filepath.Join(volume, dirUpdates, "deleted.json")

	a.skipItems = make(map[int]P)

	a.updates = newUpdates(a.updatefile)
	a.deleted = newDeleted(a.deletefile)

	// search: готовый файл индекса подключается без пересборки
	a.options.log.Println("init search...")
	a.search = search.Open(filepath.Join(volume, "search.idx"), a.stamp())

	//init
	a.init()

	if options.zip {
		a.reindexZipCodes()
	}
	a.options.log.Println("Ready!")
	return
}

type Engine struct {
	db      *store
	search  *search.DB
	options *options
	updates *updates
	deleted *deleted

	updatefile string
	deletefile string
	namesfile  string

	zipcodes map[int]*P

	mu        sync.RWMutex // skipItems
	skipItems map[int]P
	Updates   []P
}

func (a *Engine) Count() int {
	return int(a.db.Count())
}

// make defer
func (a *Engine) Close() {
	a.options.log.Println("close db...")
	a.search.Flush(a.stamp())
	a.db.Close()
}

// stamp это отпечаток всего, из чего собран индекс поиска: настройки, формат
// записи и файлы изменений. Пока он тот же, файл индекса годится как есть.
func (a *Engine) stamp() uint64 {
	h := fnv.New64a()
	fmt.Fprint(h, recVersion, a.options.hash())
	for _, name := range []string{a.updatefile, a.deletefile} {
		if st, err := os.Stat(name); err == nil {
			fmt.Fprint(h, "|", st.Size(), st.ModTime().UnixNano())
		} else {
			fmt.Fprint(h, "|-")
		}
	}
	return h.Sum64()
}

func (a *Engine) Get(id int, lang string, fields ...string) (b P) {
	b = a.db.Get(numbers.IntBytes(id))
	if b == nil {
		return nil
	}

	//origin name
	origin := b.Name()

	//name for search
	var name string
	if lang != "" {
		names := b.Names()
		if strings.TrimSpace(names[lang]) != "" {
			name = names[lang]
		} else {
			name = b.Name()
		}
	} else {
		name = b.Name()
	}

	//для админки
	if fields == nil {
		b.Set(FieldName, name)
		b.Set("origin", origin)
		return b
	}

	//для публичного поиска
	res := jsons.Creates(FieldID, id)
	for _, f := range fields {
		switch f {
		case FieldBoost:
			res.Add(FieldBoost, b.Boost())
		case FieldPeople:
			res.Add(FieldPeople, b.People())
		case FieldName:
			res.Add(FieldName, name)
		case FieldLat:
			res.Add(FieldLat, b.Lat())
		case FieldLong:
			res.Add(FieldLong, b.Long())
		case FieldCountry:
			res.Add(FieldCountry, b.Country())
		case FieldTags:
			res.Add(FieldTags, b.Tags())
		case FieldState:
			res.Add(FieldState, b.State())
		case FieldZip:
			res.Add(FieldZip, b.Zip())
		case FieldAlias:
			res.Add(FieldAlias, b.Alias())
		case FieldTZ:
			res.Add(FieldTZ, b.TZ())
		case FieldHash:
			res.Add(FieldHash, b.Hash())
		case FieldFlag:
			res.Add(FieldFlag, b.Flag())
		case FieldCur:
			res.Add(FieldCur, b.Cur())
		case FieldSymbol:
			res.Add(FieldSymbol, b.Symbol())
		case FieldLang:
			res.Add(FieldLang, b.Lang())
		case FieldS:
			res.Add(FieldS, b.S())
		case FieldNames:
			res.Add(FieldNames, b.Names())
		case "location":
			res.Add("location", Location(b.Country(), name, b.State()))
		default:
		}
	}

	return res.Bytes()
}

func Location(country, city, state string) (s string) {
	switch country {
	case "us":
		switch state {
		case "":
			return fmt.Sprintf("%s, US", city)
		default:
			return fmt.Sprintf("%s, %s", city, strings.ToTitle(state))
		}
	default:
		return fmt.Sprintf("%s, %s", city, strings.ToTitle(country))
	}
}
