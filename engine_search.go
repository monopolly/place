package place

import (
	"github.com/monopolly/place/search"
)

// Search ищет города по началу названия и возвращает json массив.
// q это текст и фильтры: "ве @ru". Пустой language определяется по запросу.
// Без fields отдаются полные записи (для админки).
func (a *Engine) Search(limit int, q string, language string, fields ...string) (json []byte) {
	if language == "" {
		language = a.options.detect(q)
	}
	lang, ok := a.options.langIndex[language]
	if !ok {
		language = "en"
		if lang, ok = a.options.langIndex[language]; !ok {
			lang = -1
		}
	}
	mask, fast := fieldMask(fields)
	slots := slotNames + len(a.options.langs)

	json = make([]byte, 0, 2+min(limit, 16)*(48+32*len(fields)))
	json = append(json, '[')
	a.search.Each(limit, q, func(x *search.Item) {
		var r rec
		if fast && r.open(x.Data, slots) {
			json = r.render(json, x.ID, lang, mask)
		} else {
			// полная запись из базы
			item := a.Get(x.ID, language, fields...)
			if item == nil {
				return
			}
			json = append(json, item...)
		}
		json = append(json, ',')
	})
	if len(json) > 1 {
		json = json[:len(json)-1]
	}
	return append(json, ']')
}
