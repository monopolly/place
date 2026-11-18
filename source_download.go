package place

import (
	"fmt"
	"os"

	"github.com/monopolly/file"
)

func (a *Engine) download(s source) (err error) {
	// download zip > volume/sourcedir
	if file.Exists(s.zipName()) {
		a.options.log.Println("source exists, skip download", s.zipName())
		return
	}

	// качаем во временный файл, чтобы оборванная загрузка не сошла за готовую
	part := s.zipName() + ".part"
	fmt.Printf("download %s...", s.name())
	r := file.DownloadFast(s.link(), part)
	err = r.Start(func(now, total int, percent float64) {
		fmt.Printf("\rdownload %s... %.2f%% %dmb/%dmb", s.name(), percent, now/1_000_000, total/1_000_000)
	})
	if err == nil {
		err = os.Rename(part, s.zipName())
	}
	if err != nil {
		fmt.Printf("\rdownload %s... %s", s.name(), err)
	}
	fmt.Println()
	return
}
