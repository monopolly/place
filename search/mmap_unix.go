//go:build unix

package search

import (
	"os"
	"syscall"
)

func mapFile(path string) (mem []byte, unmap func(), err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return
	}
	if st.Size() == 0 {
		return nil, nil, errFormat
	}
	mem, err = syscall.Mmap(int(f.Fd()), 0, int(st.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return
	}
	return mem, func() { syscall.Munmap(mem) }, nil
}
