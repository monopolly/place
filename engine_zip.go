package place

// reindex all db
func (a *Engine) reindexZipCodes() {
	a.options.log.Println("add zip codes...")

	zipIterate(func(code, city, state string, lat, long float64) {
		// var p Place

	})

}
