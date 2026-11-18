type P = {
	id              :number // int geoid
	up              :number // int64 updated
	boost           :number // int64 boost
	people          :number // int people count
	name            :string // string Moscow latin, Artplay
	lat             :number // float64 lat
	long            :number // float64 long
	hash            :string // string geohash #sfkh79247
	country         :string // string ru
	state           :string // string ca
	zip             :string // string 90210
	alias           :string // string Питер, Луки, артплей
	tz              :string // string Europe/Moscow
	flag            :string // string utf8 flag 🇺🇲
	cur             :string // string USD
	symbol          :string // string $
	lang            :string // string ru, главный язык
	s               :string // string read only search index text: new york, newyork
	o               :string[] // []string read only search options: direction:asia
	tags            :string // string #ocean #maldives через пробел
	names           :Record<string, string> // map[string]string lang translate
	del             :boolean // bool если удалено, то не входит в поиск
}