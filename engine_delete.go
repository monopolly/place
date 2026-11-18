package place

func (a *Engine) Delete(p P) {
	a.deleteFromDB(p)
	a.deleteFromSearch(p)
	a.deleted.add(p)
	a.addToSkipList(p)
	a.options.log.Println("delete", p.Flag(), p.ID(), p.Name(), p.Country())
}

func (a *Engine) DeleteUpdates(id int) {
	a.updates.delete(id)
	a.Reindex()
}

func (a *Engine) deleteFromDB(p P) {
	a.db.Delete(p.BID())
}

func (a *Engine) deleteFromSearch(p P) {
	a.search.Delete(p.ID())
}

func (a *Engine) addToSkipList(p P) {
	a.mu.Lock()
	a.skipItems[p.ID()] = p
	a.mu.Unlock()
}

func (a *Engine) skipped(id int) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.skipItems[id] != nil
}
