package place

import (
	"archive/zip"
	"bufio"
	"fmt"
)

// lines читает текстовый файл источника прямо из zip, без распаковки на диск.
// Строка действует только до возврата из h.
func (a *Engine) lines(s source, h func(line []byte)) (err error) {
	r, err := zip.OpenReader(s.zipName())
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != s.textname() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// в cities строки длинные из-за списка альтернативных названий
		scanner := bufio.NewScanner(rc)
		scanner.Buffer(make([]byte, 1<<20), 1<<20)
		for scanner.Scan() {
			h(scanner.Bytes())
		}
		return scanner.Err()
	}
	return fmt.Errorf("%s: no %s inside", s.zipName(), s.textname())
}
