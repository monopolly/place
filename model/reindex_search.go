package model

import (
	"fmt"
	"sort"
	"strings"
)

func (a *P) ReindexSearch(allowlang map[string]bool) {

	//search
	un := map[string]bool{}

	un[clear(a.Name())] = true

	for lang, x := range a.Names() {
		if allowlang != nil && !allowlang[lang] {
			continue
		}
		//Рио де жанейро > Риодежанейро
		un[clear(x)] = true
	}

	un[fmt.Sprintf("@%s", a.Hash())] = true

	list := []string{a.Alias()}
	for x := range un {
		list = append(list, x)
	}
	// порядок слов постоянный: одинаковая запись дает одинаковый json
	sort.Strings(list[1:])

	a.S(strings.Join(list, " "))
}
