package place

import "github.com/monopolly/file"

func (a *Engine) delete(s source) {
	a.options.log.Println("delete source zip", s.zipName())
	file.Delete(s.zipName())
}
