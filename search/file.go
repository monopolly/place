package search

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"
	"runtime"
	"unsafe"
)

// Файл индекса: заголовок и секции, каждая секция это один массив как есть
// в памяти. Числа в порядке байт машины, секции выровнены на 8 байт.
//
//	magic [8]byte, order uint32, sections uint32, stamp uint64,
//	sections x (offset uint64, size uint64), данные секций
const (
	fileMagic   = "plcidx01"
	fileOrder   = 0x01020304
	stampOffset = 16
	headerSize  = 24
)

var errFormat = errors.New("search: bad index file")

func bytesOf[T any](s []T) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&s[0])), len(s)*int(unsafe.Sizeof(s[0])))
}

func sliceOf[T any](b []byte) []T {
	var x T
	if len(b) < int(unsafe.Sizeof(x)) {
		return nil
	}
	return unsafe.Slice((*T)(unsafe.Pointer(&b[0])), len(b)/int(unsafe.Sizeof(x)))
}

func bytesOfString(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func stringOf(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// sections перечисляет массивы индекса в порядке записи
func (a *index) sections() [][]byte {
	return [][]byte{
		bytesOf(a.ids), bytesOf(a.boost), bytesOf(a.people), bytesOf(a.dataOff), bytesOfString(a.data),
		bytesOfString(a.terms.blob), bytesOf(a.terms.off), bytesOf(a.terms.start), bytesOf(a.terms.docs),
		bytesOfString(a.opts.blob), bytesOf(a.opts.off), bytesOf(a.opts.start), bytesOf(a.opts.docs),
		bytesOf(a.docTermStart), bytesOf(a.docTerms), bytesOf(a.docOptStart), bytesOf(a.docOpts),
		bytesOf(a.first), bytesOf(a.rmq.bmin), bytesOf(a.rmq.sp),
	}
}

const sectionCount = 20

// write пишет индекс во временный файл и подменяет им path
func (a *index) write(path string, stamp uint64) (err error) {
	f, err := os.Create(path + ".tmp")
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(path + ".tmp")
		}
	}()

	list := a.sections()
	head := make([]byte, headerSize+16*len(list))
	copy(head, fileMagic)
	binary.NativeEndian.PutUint32(head[8:], fileOrder)
	binary.NativeEndian.PutUint32(head[12:], uint32(len(list)))
	binary.NativeEndian.PutUint64(head[stampOffset:], stamp)
	pos := uint64(len(head))
	for i, s := range list {
		binary.NativeEndian.PutUint64(head[headerSize+16*i:], pos)
		binary.NativeEndian.PutUint64(head[headerSize+16*i+8:], uint64(len(s)))
		pos += uint64(len(s)+7) &^ 7
	}

	w := bufio.NewWriterSize(f, 1<<20)
	w.Write(head)
	var pad [8]byte
	for _, s := range list {
		w.Write(s)
		w.Write(pad[:(8-len(s)%8)%8])
	}
	if err = w.Flush(); err != nil {
		return
	}
	if err = f.Close(); err != nil {
		return
	}
	return os.Rename(path+".tmp", path)
}

// writeStamp меняет только метку в готовом файле
func writeStamp(path string, stamp uint64) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	var b [8]byte
	binary.NativeEndian.PutUint64(b[:], stamp)
	_, err = f.WriteAt(b[:], stampOffset)
	return err
}

// load подключает файл индекса через mmap
func load(path string) (a *index, err error) {
	mem, unmap, err := mapFile(path)
	if err != nil {
		return
	}
	defer func() {
		if err != nil && unmap != nil {
			unmap()
		}
	}()

	if len(mem) < headerSize || string(mem[:8]) != fileMagic ||
		binary.NativeEndian.Uint32(mem[8:]) != fileOrder ||
		binary.NativeEndian.Uint32(mem[12:]) != sectionCount ||
		len(mem) < headerSize+16*sectionCount {
		return nil, errFormat
	}
	var s [sectionCount][]byte
	for i := range s {
		off := binary.NativeEndian.Uint64(mem[headerSize+16*i:])
		size := binary.NativeEndian.Uint64(mem[headerSize+16*i+8:])
		if off%8 != 0 || off > uint64(len(mem)) || size > uint64(len(mem))-off {
			return nil, errFormat
		}
		s[i] = mem[off : off+size : off+size]
	}

	a = &index{stamp: binary.NativeEndian.Uint64(mem[stampOffset:]), unmap: unmap}
	a.ids, a.boost, a.people, a.dataOff, a.data = sliceOf[int64](s[0]), sliceOf[int64](s[1]), sliceOf[uint32](s[2]), sliceOf[uint32](s[3]), stringOf(s[4])
	a.terms = dict{stringOf(s[5]), sliceOf[uint32](s[6]), sliceOf[uint32](s[7]), sliceOf[uint32](s[8])}
	a.opts = dict{stringOf(s[9]), sliceOf[uint32](s[10]), sliceOf[uint32](s[11]), sliceOf[uint32](s[12])}
	a.docTermStart, a.docTerms, a.docOptStart, a.docOpts = sliceOf[uint32](s[13]), sliceOf[uint32](s[14]), sliceOf[uint32](s[15]), sliceOf[uint32](s[16])
	a.first = sliceOf[uint32](s[17])
	a.rmq.init(a.terms.docs, sliceOf[uint32](s[18]), sliceOf[uint32](s[19]))

	// размеры массивов должны сходиться, иначе файл не наш или побит
	n := len(a.ids)
	if len(a.boost) != n || len(a.people) != n || len(a.dataOff) != n+1 || int(a.dataOff[n]) != len(a.data) ||
		len(a.terms.off) < 1 || len(a.terms.start) != len(a.terms.off) || len(a.opts.off) < 1 || len(a.opts.start) != len(a.opts.off) ||
		len(a.docTermStart) != n+1 || len(a.docOptStart) != n+1 || len(a.docTerms) != len(a.terms.docs) || len(a.docOpts) != len(a.opts.docs) ||
		len(a.first) != 1<<16+1 || len(a.rmq.bmin) != len(a.terms.docs)>>blockShift || len(a.rmq.sp) != int(a.rmq.lv[len(a.rmq.lv)-1]) {
		return nil, errFormat
	}

	// память отдаем, когда индекс больше никому не нужен: поиск может идти
	// по старому снимку уже после того, как его заменили новым
	if unmap != nil {
		runtime.SetFinalizer(a, func(a *index) { a.unmap() })
	}
	return
}
