package place

// под этим ключом в базе лежит хеш настроек
const hashKey = "_"

func (a *Engine) init() {
	a.options.log.Println("init engine...")
	switch a.sameHash() {
	case true:
		a.options.log.Println("has database...")
		// deleted
		a.deleted.init()
		// changes
		a.updates.init()
		// index
		if a.search.Loaded() {
			a.options.log.Println("search index is ready...")
		} else {
			a.index()
		}
	default:
		a.options.log.Println("new database...")
		a.search.Reset()
		// city
		a.download(a.options.source)
		// names
		a.download(alternateNamesV2)
		// deleted
		a.deleted.init()
		// parse
		if err := a.parse(); err != nil {
			// хеш не сохраняем: следующий запуск попробует еще раз
			a.options.log.Println("parse error:", err)
			a.updates.init()
			a.index()
			return
		}
		// changes
		a.updates.init()
		// reindex
		a.Reindex()
		// save hash
		a.saveHash()
	}
}

// db hash vs options hash
func (a *Engine) sameHash() bool {
	return string(a.db.Get([]byte(hashKey))) == a.options.hash()
}

// set options hash
func (a *Engine) saveHash() {
	a.db.Put([]byte(hashKey), []byte(a.options.hash()))
}
