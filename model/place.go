package model

// go=Place js=P
type place struct {
	id      int               //geoid
	up      int64             //updated
	boost   int64             //boost
	people  int               //people count
	name    string            //Moscow latin, Artplay
	lat     float64           //lat
	long    float64           //long
	hash    string            //geohash #sfkh79247
	class   string            //ru
	code    string            //ru
	adm1    string            //ru
	adm2    string            //ru
	adm3    string            //ru
	adm4    string            //ru
	cc2     string            //ru
	country string            //ru
	state   string            //ca
	zip     string            //90210
	alias   string            //Питер, Луки, артплей
	tz      string            //Europe/Moscow
	flag    string            //utf8 flag 🇺🇲
	cur     string            //USD
	symbol  string            //$
	lang    string            //ru, главный язык
	s       string            //read only search index text: new york, newyork
	o       []string          //read only search options: direction:asia
	tags    string            //#ocean #maldives через пробел
	names   map[string]string //lang translate
	del     bool              //если удалено, то не входит в поиск
}
