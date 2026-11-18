package place

import (
	"github.com/monopolly/jsons"
)

func (a *Engine) Stats() []byte {
	b := jsons.Create()
	b.Add("count", a.Count())
	b.Add("allow", a.options.langmap)
	return b.Bytes()
}
