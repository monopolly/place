package place

import (
	"bytes"
)

/* convert */
/*
	7625122	3039154	link	http://en.wikipedia.org/wiki/El_Tarter
	7911398	3039154	ru	Эл Тартер	1
	1297837	3039155	ca	Sant Miquel d'Engolasters	1
	1297838	3039156		Sant Miquel
	4307204	3039156		Sant Miquel de Fontaneda
*/

/* func main() {

	//list := make(map[int][]string)
	//file.LoadJson("names.json", &list)
	//fmt.Println("done")
	convertNames()
	//select {}
	time.Sleep(time.Second * 60)
} */

/*
	0 alternateNameId   : the id of this alternate name, int
	1 geonameid         : geonameId referring to id in table 'geoname', int
	2 isolanguage       : iso 639 language code 2- or 3-characters; 4-characters 'post' for postal codes and 'iata','icao' and faac for airport codes, fr_1793 for French Revolution names,  abbr for abbreviation, link to a website (mostly to wikipedia), wkdt for the wikidataid, varchar(7)
	3 alternate name    : alternate name or name variant, varchar(400)
	4 isPreferredName   : '1', if this alternate name is an official/preferred name
	5 isShortName       : '1', if this is a short name like 'California' for 'State of California'
	6 isColloquial      : '1', if this alternate name is a colloquial or slang term. Example: 'Big Apple' for 'New York'.
	7 isHistoric        : '1', if this alternate name is historic and was used in the past. Example 'Bombay' for 'Mumbai'.
	8 from		  		: from period when the name was used
	9 to		  		: to period when the name was used
*/

// alternateNamesV2.txt: в файле названия всех 12 млн объектов geonames,
// поэтому оставляем только строки нужных городов. Названия города копятся
// одной строкой байт (см. decodeNames): это в разы меньше, чем map на город.
func (a *Engine) parseNames(src source, index map[int]int32) (names [][]byte, err error) {
	names = make([][]byte, len(index))
	var count int
	err = a.lines(src, func(line []byte) {
		var p [10][]byte
		if !splitTabs(line, &p) {
			return
		}

		lang, name := bytes.TrimSpace(p[2]), p[3]
		switch {
		case len(lang) == 0 || len(lang) > 2: //link, iata
			return
		case len(p[6]) == 1 && p[6][0] == '1': //big apple
			return
		case len(p[7]) == 1 && p[7][0] == '1': //bombay old name
			return
		case len(name) > 0xffff:
			return
		}

		id, ok := atoi(p[1])
		if !ok || id == 0 {
			return
		}
		i, ok := index[id]
		if !ok {
			return
		}

		// запись: официальное ли (isPreferredName), длина языка, язык, длина названия, название
		var pref byte
		if len(p[4]) == 1 && p[4][0] == '1' {
			pref = 1
		}
		b := append(names[i], pref, byte(len(lang)))
		b = append(b, lang...)
		b = append(b, byte(len(name)), byte(len(name)>>8))
		names[i] = append(b, name...)
		count++
	})
	a.options.log.Println("names:", count)
	return
}

// decodeNames разворачивает названия города в язык > название. Из нескольких
// названий на одном языке остается последнее, но официальное не перебивается.
func decodeNames(b []byte) (names map[string]string) {
	if len(b) == 0 {
		return
	}
	names = make(map[string]string)
	preferred := make(map[string]bool)
	for len(b) >= 2 {
		pref, n := b[0] == 1, int(b[1])
		lang := string(b[2 : 2+n])
		b = b[2+n:]
		n = int(b[0]) | int(b[1])<<8
		name := string(b[2 : 2+n])
		b = b[2+n:]
		if preferred[lang] && !pref {
			continue
		}
		if pref {
			preferred[lang] = true
		}
		names[lang] = name
	}
	return
}

// splitTabs режет строку ровно на 10 полей
func splitTabs(line []byte, p *[10][]byte) bool {
	for i := 0; i < 9; i++ {
		n := bytes.IndexByte(line, '\t')
		if n < 0 {
			return false
		}
		p[i], line = line[:n], line[n+1:]
	}
	p[9] = line
	return bytes.IndexByte(line, '\t') < 0
}

func atoi(b []byte) (n int, ok bool) {
	if len(b) == 0 || len(b) > 18 {
		return
	}
	for _, c := range b {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}
