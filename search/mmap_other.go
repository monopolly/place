//go:build !unix

package search

import "os"

// без mmap файл просто читается в память
func mapFile(path string) (mem []byte, unmap func(), err error) {
	mem, err = os.ReadFile(path)
	return
}
