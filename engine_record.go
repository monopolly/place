package place

import (
	"encoding/json"
	"strconv"
	"strings"
	"unsafe"

	"github.com/monopolly/place/internal/geo"
	"github.com/monopolly/place/search"
	"github.com/tidwall/gjson"
)

// Запись города для выдачи поиска это готовые куски json одной строкой: сначала
// длины кусков по байту на каждый, дальше сами значения. Запись лежит прямо
// в индексе, поэтому поиск не ходит в базу и не разбирает json.
//
// Кусок нулевой длины значит "как обычно": пустая строка, ноль или то, что
// вычисляется из остальных полей (флаг и валюта из страны, geohash из
// координат). Так в записи остается только то, что нельзя вывести.
type rec struct {
	data string
	off  [maxSlots + 1]uint16
}

// больше кусков в запись не помещаем (16 полей + языки)
const maxSlots = 64

// recVersion входит в метку индекса: при смене формата записи файл пересобирается
const recVersion = "rec2"

// куски записи
const (
	slotAlias = iota
	slotBoost
	slotCountry
	slotCur
	slotFlag
	slotHash
	slotLang
	slotLat
	slotLong
	slotName
	slotPeople
	slotState
	slotSymbol
	slotTags
	slotTZ
	slotZip
	slotLocation // хвост location без кавычек: `, RU`
	slotNames    // дальше названия по языкам в порядке options.langs

	// не куски, а особые поля выдачи
	slotID       = -1
	slotFullName = -2
	slotFullLoc  = -3
)

// поля выдачи в том порядке, в каком они идут в json (по алфавиту)
var outFields = [...]struct {
	key  string
	slot int
}{
	{`"alias":`, slotAlias},
	{`"boost":`, slotBoost},
	{`"country":`, slotCountry},
	{`"cur":`, slotCur},
	{`"flag":`, slotFlag},
	{`"hash":`, slotHash},
	{`"id":`, slotID},
	{`"lang":`, slotLang},
	{`"lat":`, slotLat},
	{`"location":`, slotFullLoc},
	{`"long":`, slotLong},
	{`"name":`, slotFullName},
	{`"people":`, slotPeople},
	{`"state":`, slotState},
	{`"symbol":`, slotSymbol},
	{`"tags":`, slotTags},
	{`"tz":`, slotTZ},
	{`"zip":`, slotZip},
}

const maskID = 1 << 6

// fieldMask переводит список полей в биты outFields. fast == false значит,
// что нужна полная запись из базы: полей нет вовсе (админка) или просят s и names.
func fieldMask(fields []string) (mask uint32, fast bool) {
	mask, fast = maskID, len(fields) > 0
	for _, f := range fields {
		switch f {
		case FieldAlias:
			mask |= 1 << 0
		case FieldBoost:
			mask |= 1 << 1
		case FieldCountry:
			mask |= 1 << 2
		case FieldCur:
			mask |= 1 << 3
		case FieldFlag:
			mask |= 1 << 4
		case FieldHash:
			mask |= 1 << 5
		case FieldLang:
			mask |= 1 << 7
		case FieldLat:
			mask |= 1 << 8
		case "location":
			mask |= 1 << 9
		case FieldLong:
			mask |= 1 << 10
		case FieldName:
			mask |= 1 << 11
		case FieldPeople:
			mask |= 1 << 12
		case FieldState:
			mask |= 1 << 13
		case FieldSymbol:
			mask |= 1 << 14
		case FieldTags:
			mask |= 1 << 15
		case FieldTZ:
			mask |= 1 << 16
		case FieldZip:
			mask |= 1 << 17
		case FieldS, FieldNames:
			fast = false
		}
	}
	return
}

// packRec собирает запись из кусков, пустая строка если не помещается
func packRec(slots []string) string {
	size := len(slots)
	for _, s := range slots {
		if len(s) > 255 {
			return ""
		}
		size += len(s)
	}
	b := make([]byte, len(slots), size)
	for i, s := range slots {
		b[i] = byte(len(s))
		b = append(b, s...)
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// open размечает запись из n кусков
func (r *rec) open(data string, n int) bool {
	if n > maxSlots || len(data) < n {
		return false
	}
	pos := n
	for i := 0; i < n; i++ {
		r.off[i] = uint16(pos)
		pos += int(data[i])
	}
	r.off[n] = uint16(pos)
	r.data = data
	return pos == len(data)
}

func (r *rec) slot(i int) string {
	return r.data[r.off[i]:r.off[i+1]]
}

// inner снимает кавычки с json строки
func inner(raw string) string {
	if len(raw) < 2 {
		return ""
	}
	return raw[1 : len(raw)-1]
}

// appendName дописывает название на языке lang (номер в options.langs) или основное
func (r *rec) appendName(dst []byte, lang int) []byte {
	if lang >= 0 {
		if s := r.slot(slotNames + lang); s != "" {
			return append(dst, inner(s)...)
		}
	}
	return append(dst, inner(r.slot(slotName))...)
}

// appendLocation дописывает хвост location: ", NY" для штатов, иначе ", RU"
func appendLocation(dst []byte, country, state string) []byte {
	x := inner(country)
	if x == "us" {
		if s := inner(state); s != "" {
			x = s
		}
	}
	dst = append(dst, ", "...)
	for i := 0; i < len(x); i++ {
		c := x[i]
		if 'a' <= c && c <= 'z' {
			c -= 'a' - 'A'
		}
		dst = append(dst, c)
	}
	return dst
}

func appendHash(dst []byte, lat, long string) []byte {
	x, _ := strconv.ParseFloat(lat, 64)
	y, _ := strconv.ParseFloat(long, 64)
	dst = append(dst, '"')
	dst = geo.AppendGeohash(dst, x, y)
	return append(dst, '"')
}

// appendDefault дописывает значение куска, которого в записи нет
func (r *rec) appendDefault(dst []byte, slot int) []byte {
	switch slot {
	case slotBoost, slotLat, slotLong, slotPeople:
		return append(dst, '0')
	case slotHash:
		return appendHash(dst, r.slot(slotLat), r.slot(slotLong))
	case slotLocation:
		return appendLocation(dst, r.slot(slotCountry), r.slot(slotState))
	}
	var s string
	switch slot {
	case slotCur:
		s = geo.Lookup(inner(r.slot(slotCountry))).Cur
	case slotFlag:
		s = geo.Lookup(inner(r.slot(slotCountry))).Flag
	case slotLang:
		s = geo.Lookup(inner(r.slot(slotCountry))).Lang
	case slotSymbol:
		s = geo.Lookup(inner(r.slot(slotCountry))).Symbol
	}
	dst = append(dst, '"')
	dst = append(dst, s...)
	return append(dst, '"')
}

func (r *rec) appendSlot(dst []byte, slot int) []byte {
	if s := r.slot(slot); s != "" {
		return append(dst, s...)
	}
	return r.appendDefault(dst, slot)
}

// render дописывает json города с полями из mask
func (r *rec) render(dst []byte, id int, lang int, mask uint32) []byte {
	dst = append(dst, '{')
	for i := range outFields {
		if mask&(1<<i) == 0 {
			continue
		}
		f := &outFields[i]
		dst = append(dst, f.key...)
		switch f.slot {
		case slotID:
			dst = strconv.AppendInt(dst, int64(id), 10)
		case slotFullName:
			dst = append(dst, '"')
			dst = r.appendName(dst, lang)
			dst = append(dst, '"')
		case slotFullLoc:
			dst = append(dst, '"')
			dst = r.appendName(dst, lang)
			dst = r.appendSlot(dst, slotLocation)
			dst = append(dst, '"')
		default:
			dst = r.appendSlot(dst, f.slot)
		}
		dst = append(dst, ',')
	}
	dst[len(dst)-1] = '}'
	return dst
}

// parsePlace за один проход по json достает все, что нужно индексу
func (a *Engine) parsePlace(p P) (item search.Item, text string, options []string) {
	langs := a.options.langIndex
	slots := make([]string, slotNames+len(langs))
	var country, state string

	str := func(v gjson.Result) string {
		if v.Type == gjson.String {
			return v.Raw
		}
		return ""
	}
	// без копии: все, что остается жить после разбора, копируется ниже
	gjson.Parse(unsafe.String(unsafe.SliceData(p), len(p))).ForEach(func(k, v gjson.Result) bool {
		switch k.Str {
		case FieldID:
			item.ID = int(v.Int())
		case FieldPeople:
			item.People = int(v.Int())
			slots[slotPeople] = strconv.Itoa(item.People)
		case FieldBoost:
			item.Boost = v.Int()
			slots[slotBoost] = strconv.FormatInt(item.Boost, 10)
		case FieldS:
			text = v.String() // search режет на слова и копирует
		case FieldO:
			v.ForEach(func(_, o gjson.Result) bool {
				options = append(options, o.String())
				return true
			})
		case FieldNames:
			v.ForEach(func(l, name gjson.Result) bool {
				if i, ok := langs[l.Str]; ok && name.Type == gjson.String && strings.TrimSpace(name.Str) != "" {
					slots[slotNames+i] = name.Raw
				}
				return true
			})
		case FieldLat:
			if v.Type == gjson.Number {
				slots[slotLat] = v.Raw
			}
		case FieldLong:
			if v.Type == gjson.Number {
				slots[slotLong] = v.Raw
			}
		case FieldCountry:
			country, slots[slotCountry] = v.String(), str(v)
		case FieldState:
			state, slots[slotState] = v.String(), str(v)
		case FieldName:
			slots[slotName] = str(v)
		case FieldAlias:
			slots[slotAlias] = str(v)
		case FieldCur:
			slots[slotCur] = str(v)
		case FieldFlag:
			slots[slotFlag] = str(v)
		case FieldHash:
			slots[slotHash] = str(v)
		case FieldLang:
			slots[slotLang] = str(v)
		case FieldSymbol:
			slots[slotSymbol] = str(v)
		case FieldTags:
			slots[slotTags] = str(v)
		case FieldTZ:
			slots[slotTZ] = str(v)
		case FieldZip:
			slots[slotZip] = str(v)
		}
		return true
	})
	slots[slotLocation] = jsonInner(Location(country, "", state))

	// выкидываем все, что при выдаче получится само; r смотрит в полную запись,
	// так что вычисляемые значения считаются по исходным полям
	var r rec
	var buf [64]byte
	if len(slots) <= maxSlots && r.open(packRec(slots), len(slots)) {
		for i := 0; i < slotNames; i++ {
			if slots[i] == string(r.appendDefault(buf[:0], i)) {
				slots[i] = ""
			}
		}
		item.Data = packRec(slots)
	}
	return
}

// jsonInner это строка в виде json без кавычек
func jsonInner(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}
