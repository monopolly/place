package model

import (
	"fmt"
	"reflect"
	"time"

	"github.com/monopolly/cast"
	"github.com/monopolly/jsons"
)

/*
	Model Generator
	Help to auto-generated correct models for golang structures.
	Martin Prestone (c) 2024
	github.com/monopolly
	Ex: //! [] go=News js=NewsJson sql=news noinit up...

	STRUCT:
	go=News        Set golang struct names. Ex: go=News1 > type News1 struct{}
	ts=news        Set typescript struct names. Ex: ts=NewsJson
	swift          Generate swift model
	enum           Generate swift enum
	sql=news       Set sql table name. Ex: sql=accounts
	noinit         No init function for struct
	js=NewsJson    Change json struct names. Ex: js=NewsJson
	demo           Generate demo json file with default values
	noprefix       Generate simple index IndexID instead IndexNewsID
	lock!          Lock model for new jsons generation. Can't change model.
	!omit          No omit tag for json

	Field:
	title{}        Add custom title for index. title{Nice & Sweet}
	desc{}         Add custom desc for index. desc{This field for success}

	GO:
	up             Make uppercase for functions
	type=""        Replace golang type. Ex: type="[]*News"
	name=""        Replace golang struct name. Ex: name="NewsList"
	title=""       Add custom title for index
	desc=""        Add custom desc for index
	title{}        Add field user title title{Nice}
	desc{}         Add field user title desc{Use it nice}
	nofunc         Do not create any jsons functions for fiels
	must           Create one validation function for all must fields
	req{}          Add user required fields
	#              Add lists for fields. Ex: id int //#readonly #must...

	SQL:
	index          Create simple default index or gin for jsonb
	defaults       Add default value based on field type
	skip           Skip field for sql queries
	type=""        Rewrite sql type for field. Ex: type="jsonb"
	add=""         Append sql for field. Ex: replace="primary key"
	noinsert       Do use field for insert function
	idx            Add index fields by group name. Ex: idx="nameIndex" idx="credsIndex"
	search         Add tsvector index by group. Ex: search="tsv": tsv tsvector GENERATED ALWAYS AS (to_tsvector('simple', title || ' ' || brand)). For search: SELECT brand, title FROM assets WHERE search @@ to_tsquery('english', 'f8');
	unix           Add default value: extract(epoch from now())
	replace=""     Rewrite sql for field. Ex: replace="bigint primary key"
	unique=""      Add unique fields constrains by group (you need set group name, then generator join fields). Ex: unique="group1" unique="group2"

	SWIFT:
	must           Swift field with required values
	skip           Ignore field for Swift
	type=""        Replace type for Swift model. Ex: type="string"
	file=""        Create another swift file for this field Swift model. file="f1", file="f2"

	JSON:
	raw            Set Raw json function inside field
	name=""        Replace json field name. Ex: name="sid"
	skip           Skip json field
	inc            Add inc jsons function for numbers fields
	bool           Add set jsons function for bool fields
	time           Create convert jsons function for unixtime fields
*/

// порядковые номера полей
type PlaceIndexType int

const (
	IndexPlaceID = PlaceIndexType(iota)
	IndexPlaceUP
	IndexPlaceBoost
	IndexPlacePeople
	IndexPlaceName
	IndexPlaceLat
	IndexPlaceLong
	IndexPlaceHash
	IndexPlaceCountry
	IndexPlaceState
	IndexPlaceZip
	IndexPlaceAlias
	IndexPlaceTZ
	IndexPlaceFlag
	IndexPlaceCur
	IndexPlaceSymbol
	IndexPlaceLang
	IndexPlaceS
	IndexPlaceO
	IndexPlaceTags
	IndexPlaceNames
	IndexPlaceDel
)

const (
	FieldPlaceID      = "id"      // int geoid
	FieldPlaceUP      = "up"      // int64 updated
	FieldPlaceBoost   = "boost"   // int64 boost
	FieldPlacePeople  = "people"  // int people count
	FieldPlaceName    = "name"    // string Moscow latin, Artplay
	FieldPlaceLat     = "lat"     // float64 lat
	FieldPlaceLong    = "long"    // float64 long
	FieldPlaceHash    = "hash"    // string geohash #sfkh79247
	FieldPlaceCountry = "country" // string ru
	FieldPlaceState   = "state"   // string ca
	FieldPlaceZip     = "zip"     // string 90210
	FieldPlaceAlias   = "alias"   // string Питер, Луки, артплей
	FieldPlaceTZ      = "tz"      // string Europe/Moscow
	FieldPlaceFlag    = "flag"    // string utf8 flag 🇺🇲
	FieldPlaceCur     = "cur"     // string USD
	FieldPlaceSymbol  = "symbol"  // string $
	FieldPlaceLang    = "lang"    // string ru, главный язык
	FieldPlaceS       = "s"       // string read only search index text: new york, newyork
	FieldPlaceO       = "o"       // []string read only search options: direction:asia
	FieldPlaceTags    = "tags"    // string #ocean #maldives через пробел
	FieldPlaceNames   = "names"   // map[string]string lang translate
	FieldPlaceDel     = "del"     // bool если удалено, то не входит в поиск
)

func PlaceIndexes() []PlaceIndexType {
	return []PlaceIndexType{IndexPlaceID, IndexPlaceUP, IndexPlaceBoost, IndexPlacePeople, IndexPlaceName, IndexPlaceLat, IndexPlaceLong, IndexPlaceHash, IndexPlaceCountry, IndexPlaceState, IndexPlaceZip, IndexPlaceAlias, IndexPlaceTZ, IndexPlaceFlag, IndexPlaceCur, IndexPlaceSymbol, IndexPlaceLang, IndexPlaceS, IndexPlaceO, IndexPlaceTags, IndexPlaceNames, IndexPlaceDel}

}

// public struct with json tags
type Place struct {
	ID      int               `json:"id,omitempty"`      // geoid
	UP      int64             `json:"up,omitempty"`      // updated
	Boost   int64             `json:"boost,omitempty"`   // boost
	People  int               `json:"people,omitempty"`  // people count
	Name    string            `json:"name,omitempty"`    // Moscow latin, Artplay
	Lat     float64           `json:"lat,omitempty"`     // lat
	Long    float64           `json:"long,omitempty"`    // long
	Hash    string            `json:"hash,omitempty"`    // geohash #sfkh79247
	Country string            `json:"country,omitempty"` // ru
	State   string            `json:"state,omitempty"`   // ca
	Zip     string            `json:"zip,omitempty"`     // 90210
	Alias   string            `json:"alias,omitempty"`   // Питер, Луки, артплей
	TZ      string            `json:"tz,omitempty"`      // Europe/Moscow
	Flag    string            `json:"flag,omitempty"`    // utf8 flag 🇺🇲
	Cur     string            `json:"cur,omitempty"`     // USD
	Symbol  string            `json:"symbol,omitempty"`  // $
	Lang    string            `json:"lang,omitempty"`    // ru, главный язык
	S       string            `json:"s,omitempty"`       // read only search index text: new york, newyork
	O       []string          `json:"o,omitempty"`       // read only search options: direction:asia
	Tags    string            `json:"tags,omitempty"`    // #ocean #maldives через пробел
	Names   map[string]string `json:"names,omitempty"`   // lang translate
	Del     bool              `json:"del,omitempty"`     // если удалено, то не входит в поиск
}

// Unmarshal json to struct
func JsonToStructPlace(r []byte) (a Place, err error) {
	err = jsons.Unmarshal(r, &a)
	return
}

// Parse []any to ID struct
func ParsePlaceToStruct(r []any) (a *Place) {
	a = new(Place)

	for pos, x := range r {
		switch PlaceIndexType(pos) {
		case IndexPlaceID:
			a.ID = cast.Int(x) //int
		case IndexPlaceUP:
			a.UP = cast.Int64(x) //int64
		case IndexPlaceBoost:
			a.Boost = cast.Int64(x) //int64
		case IndexPlacePeople:
			a.People = cast.Int(x) //int
		case IndexPlaceName:
			a.Name = cast.String(x) //string
		case IndexPlaceLat:
			a.Lat = cast.Float(x) //float64
		case IndexPlaceLong:
			a.Long = cast.Float(x) //float64
		case IndexPlaceHash:
			a.Hash = cast.String(x) //string
		case IndexPlaceCountry:
			a.Country = cast.String(x) //string
		case IndexPlaceState:
			a.State = cast.String(x) //string
		case IndexPlaceZip:
			a.Zip = cast.String(x) //string
		case IndexPlaceAlias:
			a.Alias = cast.String(x) //string
		case IndexPlaceTZ:
			a.TZ = cast.String(x) //string
		case IndexPlaceFlag:
			a.Flag = cast.String(x) //string
		case IndexPlaceCur:
			a.Cur = cast.String(x) //string
		case IndexPlaceSymbol:
			a.Symbol = cast.String(x) //string
		case IndexPlaceLang:
			a.Lang = cast.String(x) //string
		case IndexPlaceS:
			a.S = cast.String(x) //string
		case IndexPlaceO:
			a.O = cast.SliceString(x) //[]string
		case IndexPlaceTags:
			a.Tags = cast.String(x) //string
		case IndexPlaceNames:
			a.Names = cast.StringMapString(x) //map[string]string
		case IndexPlaceDel:
			a.Del = cast.Bool(x) //bool
		}
	}
	return
}

// Tuple create an array from struct
func (a *Place) Tuple() (r []any) {
	return []any{a.ID, a.UP, a.Boost, a.People, a.Name, a.Lat, a.Long, a.Hash, a.Country, a.State, a.Zip, a.Alias, a.TZ, a.Flag, a.Cur, a.Symbol, a.Lang, a.S, a.O, a.Tags, a.Names, a.Del}
}

// Tuple create an array from struct
func (a *Place) sqlTuple() (r []any) {
	return []any{a.ID, a.UP, a.Boost, a.People, a.Name, a.Lat, a.Long, a.Hash, a.Country, a.State, a.Zip, a.Alias, a.TZ, a.Flag, a.Cur, a.Symbol, a.Lang, a.S, a.O, a.Tags, a.Names, a.Del}
}

// update struct with function
func (a *Place) Update(k string, x any) {
	switch k {
	case "id":
		a.ID = cast.Int(x) //int
	case "up":
		a.UP = cast.Int64(x) //int64
	case "boost":
		a.Boost = cast.Int64(x) //int64
	case "people":
		a.People = cast.Int(x) //int
	case "name":
		a.Name = cast.String(x) //string
	case "lat":
		a.Lat = cast.Float(x) //float64
	case "long":
		a.Long = cast.Float(x) //float64
	case "hash":
		a.Hash = cast.String(x) //string
	case "country":
		a.Country = cast.String(x) //string
	case "state":
		a.State = cast.String(x) //string
	case "zip":
		a.Zip = cast.String(x) //string
	case "alias":
		a.Alias = cast.String(x) //string
	case "tz":
		a.TZ = cast.String(x) //string
	case "flag":
		a.Flag = cast.String(x) //string
	case "cur":
		a.Cur = cast.String(x) //string
	case "symbol":
		a.Symbol = cast.String(x) //string
	case "lang":
		a.Lang = cast.String(x) //string
	case "s":
		a.S = cast.String(x) //string
	case "o":
		a.O = cast.SliceString(x) //[]string
	case "tags":
		a.Tags = cast.String(x) //string
	case "names":
		a.Names = cast.StringMapString(x) //map[string]string
	case "del":
		a.Del = cast.Bool(x) //bool
	}
	return
}

// get struct value with function
func (a *Place) Get(k string) (v any) {
	switch k {
	case "id":
		return a.ID //int
	case "up":
		return a.UP //int64
	case "boost":
		return a.Boost //int64
	case "people":
		return a.People //int
	case "name":
		return a.Name //string
	case "lat":
		return a.Lat //float64
	case "long":
		return a.Long //float64
	case "hash":
		return a.Hash //string
	case "country":
		return a.Country //string
	case "state":
		return a.State //string
	case "zip":
		return a.Zip //string
	case "alias":
		return a.Alias //string
	case "tz":
		return a.TZ //string
	case "flag":
		return a.Flag //string
	case "cur":
		return a.Cur //string
	case "symbol":
		return a.Symbol //string
	case "lang":
		return a.Lang //string
	case "s":
		return a.S //string
	case "o":
		return a.O //[]string
	case "tags":
		return a.Tags //string
	case "names":
		return a.Names //map[string]string
	case "del":
		return a.Del //bool
	}
	return
}

// get any struct value as string
func (a *Place) String(k string) (v string) {
	switch k {
	case "id":
		return fmt.Sprint(a.ID) //int
	case "up":
		return fmt.Sprint(a.UP) //int64
	case "boost":
		return fmt.Sprint(a.Boost) //int64
	case "people":
		return fmt.Sprint(a.People) //int
	case "name":
		return fmt.Sprint(a.Name) //string
	case "lat":
		return fmt.Sprint(a.Lat) //float64
	case "long":
		return fmt.Sprint(a.Long) //float64
	case "hash":
		return fmt.Sprint(a.Hash) //string
	case "country":
		return fmt.Sprint(a.Country) //string
	case "state":
		return fmt.Sprint(a.State) //string
	case "zip":
		return fmt.Sprint(a.Zip) //string
	case "alias":
		return fmt.Sprint(a.Alias) //string
	case "tz":
		return fmt.Sprint(a.TZ) //string
	case "flag":
		return fmt.Sprint(a.Flag) //string
	case "cur":
		return fmt.Sprint(a.Cur) //string
	case "symbol":
		return fmt.Sprint(a.Symbol) //string
	case "lang":
		return fmt.Sprint(a.Lang) //string
	case "s":
		return fmt.Sprint(a.S) //string
	case "o":
		return fmt.Sprint(a.O) //[]string
	case "tags":
		return fmt.Sprint(a.Tags) //string
	case "names":
		return fmt.Sprint(a.Names) //map[string]string
	case "del":
		return fmt.Sprint(a.Del) //bool
	}
	return
}

// Struct to json
func (a *Place) ToJson() (r []byte) {
	js := jsons.Create().
		Add(FieldPlaceID, a.ID).
		Add(FieldPlaceUP, a.UP).
		Add(FieldPlaceBoost, a.Boost).
		Add(FieldPlacePeople, a.People).
		Add(FieldPlaceName, a.Name).
		Add(FieldPlaceLat, a.Lat).
		Add(FieldPlaceLong, a.Long).
		Add(FieldPlaceHash, a.Hash).
		Add(FieldPlaceCountry, a.Country).
		Add(FieldPlaceState, a.State).
		Add(FieldPlaceZip, a.Zip).
		Add(FieldPlaceAlias, a.Alias).
		Add(FieldPlaceTZ, a.TZ).
		Add(FieldPlaceFlag, a.Flag).
		Add(FieldPlaceCur, a.Cur).
		Add(FieldPlaceSymbol, a.Symbol).
		Add(FieldPlaceLang, a.Lang).
		Add(FieldPlaceS, a.S).
		Add(FieldPlaceO, a.O).
		Add(FieldPlaceTags, a.Tags).
		Add(FieldPlaceNames, a.Names).
		Add(FieldPlaceDel, a.Del)
	return js.Bytes()
}

func PlaceOceanList() []PlaceIndexType {
	return []PlaceIndexType{IndexPlaceTags}
}

func PlaceMaldivesList() []PlaceIndexType {
	return []PlaceIndexType{IndexPlaceTags}
}

func PlaceSfkh79247List() []PlaceIndexType {
	return []PlaceIndexType{IndexPlaceHash}
}

func (a PlaceIndexType) Sfkh79247() bool {
	switch a {
	case IndexPlaceHash:
		return true
	default:
		return false
	}
}
func (a PlaceIndexType) Ocean() bool {
	switch a {
	case IndexPlaceTags:
		return true
	default:
		return false
	}
}
func (a PlaceIndexType) Maldives() bool {
	switch a {
	case IndexPlaceTags:
		return true
	default:
		return false
	}
}

// key index string
func (a PlaceIndexType) String() string {
	switch a {
	case IndexPlaceID:
		return "id"
	case IndexPlaceUP:
		return "up"
	case IndexPlaceBoost:
		return "boost"
	case IndexPlacePeople:
		return "people"
	case IndexPlaceName:
		return "name"
	case IndexPlaceLat:
		return "lat"
	case IndexPlaceLong:
		return "long"
	case IndexPlaceHash:
		return "hash"
	case IndexPlaceCountry:
		return "country"
	case IndexPlaceState:
		return "state"
	case IndexPlaceZip:
		return "zip"
	case IndexPlaceAlias:
		return "alias"
	case IndexPlaceTZ:
		return "tz"
	case IndexPlaceFlag:
		return "flag"
	case IndexPlaceCur:
		return "cur"
	case IndexPlaceSymbol:
		return "symbol"
	case IndexPlaceLang:
		return "lang"
	case IndexPlaceS:
		return "s"
	case IndexPlaceO:
		return "o"
	case IndexPlaceTags:
		return "tags"
	case IndexPlaceNames:
		return "names"
	case IndexPlaceDel:
		return "del"
	default:
		return ""
	}
}

// key index string
func (a PlaceIndexType) SQLName() string {
	switch a {
	case IndexPlaceID:
		return "id"
	case IndexPlaceUP:
		return "up"
	case IndexPlaceBoost:
		return "boost"
	case IndexPlacePeople:
		return "people"
	case IndexPlaceName:
		return "name"
	case IndexPlaceLat:
		return "lat"
	case IndexPlaceLong:
		return "long"
	case IndexPlaceHash:
		return "hash"
	case IndexPlaceCountry:
		return "country"
	case IndexPlaceState:
		return "state"
	case IndexPlaceZip:
		return "zip"
	case IndexPlaceAlias:
		return "alias"
	case IndexPlaceTZ:
		return "tz"
	case IndexPlaceFlag:
		return "flag"
	case IndexPlaceCur:
		return "cur"
	case IndexPlaceSymbol:
		return "symbol"
	case IndexPlaceLang:
		return "lang"
	case IndexPlaceS:
		return "s"
	case IndexPlaceO:
		return "o"
	case IndexPlaceTags:
		return "tags"
	case IndexPlaceNames:
		return "names"
	case IndexPlaceDel:
		return "del"
	default:
		return ""
	}
}

// key index type
func (a PlaceIndexType) Type() string {
	switch a {
	case IndexPlaceID:
		return "int"
	case IndexPlaceUP:
		return "int64"
	case IndexPlaceBoost:
		return "int64"
	case IndexPlacePeople:
		return "int"
	case IndexPlaceName:
		return "string"
	case IndexPlaceLat:
		return "float64"
	case IndexPlaceLong:
		return "float64"
	case IndexPlaceHash:
		return "string"
	case IndexPlaceCountry:
		return "string"
	case IndexPlaceState:
		return "string"
	case IndexPlaceZip:
		return "string"
	case IndexPlaceAlias:
		return "string"
	case IndexPlaceTZ:
		return "string"
	case IndexPlaceFlag:
		return "string"
	case IndexPlaceCur:
		return "string"
	case IndexPlaceSymbol:
		return "string"
	case IndexPlaceLang:
		return "string"
	case IndexPlaceS:
		return "string"
	case IndexPlaceO:
		return "[]string"
	case IndexPlaceTags:
		return "string"
	case IndexPlaceNames:
		return "map[string]string"
	case IndexPlaceDel:
		return "bool"
	default:
		return ""
	}
}

// custom title
func (a PlaceIndexType) Title() string {
	switch a {
	case IndexPlaceID:
		return "ID"
	case IndexPlaceUP:
		return "UP"
	case IndexPlaceBoost:
		return "Boost"
	case IndexPlacePeople:
		return "People"
	case IndexPlaceName:
		return "Name"
	case IndexPlaceLat:
		return "Lat"
	case IndexPlaceLong:
		return "Long"
	case IndexPlaceHash:
		return "Hash"
	case IndexPlaceCountry:
		return "Country"
	case IndexPlaceState:
		return "State"
	case IndexPlaceZip:
		return "Zip"
	case IndexPlaceAlias:
		return "Alias"
	case IndexPlaceTZ:
		return "TZ"
	case IndexPlaceFlag:
		return "Flag"
	case IndexPlaceCur:
		return "Cur"
	case IndexPlaceSymbol:
		return "Symbol"
	case IndexPlaceLang:
		return "Lang"
	case IndexPlaceS:
		return "S"
	case IndexPlaceO:
		return "O"
	case IndexPlaceTags:
		return "Tags"
	case IndexPlaceNames:
		return "Names"
	case IndexPlaceDel:
		return "Del"
	default:
		return ""
	}
}

// custom desc
func (a PlaceIndexType) Desc() string {
	switch a {
	default:
		return ""
	}
}

// struct key to index
func PlaceKeyIndex(key string) PlaceIndexType {
	switch key {
	case "id":
		return IndexPlaceID
	case "up":
		return IndexPlaceUP
	case "boost":
		return IndexPlaceBoost
	case "people":
		return IndexPlacePeople
	case "name":
		return IndexPlaceName
	case "lat":
		return IndexPlaceLat
	case "long":
		return IndexPlaceLong
	case "hash":
		return IndexPlaceHash
	case "country":
		return IndexPlaceCountry
	case "state":
		return IndexPlaceState
	case "zip":
		return IndexPlaceZip
	case "alias":
		return IndexPlaceAlias
	case "tz":
		return IndexPlaceTZ
	case "flag":
		return IndexPlaceFlag
	case "cur":
		return IndexPlaceCur
	case "symbol":
		return IndexPlaceSymbol
	case "lang":
		return IndexPlaceLang
	case "s":
		return IndexPlaceS
	case "o":
		return IndexPlaceO
	case "tags":
		return IndexPlaceTags
	case "names":
		return IndexPlaceNames
	case "del":
		return IndexPlaceDel
	default:
		return 0
	}
}

// valid struct key check
func PlaceValidKey(key string) bool {
	switch key {
	case "id", "up", "boost", "people", "name", "lat", "long", "hash", "country", "state", "zip", "alias", "tz", "flag", "cur", "symbol", "lang", "s", "o", "tags", "names", "del":
		return true
	default:
		return false
	}
}

// struct to map
func (a *Place) Map() map[string]any {
	return map[string]any{
		"id":      a.ID,
		"up":      a.UP,
		"boost":   a.Boost,
		"people":  a.People,
		"name":    a.Name,
		"lat":     a.Lat,
		"long":    a.Long,
		"hash":    a.Hash,
		"country": a.Country,
		"state":   a.State,
		"zip":     a.Zip,
		"alias":   a.Alias,
		"tz":      a.TZ,
		"flag":    a.Flag,
		"cur":     a.Cur,
		"symbol":  a.Symbol,
		"lang":    a.Lang,
		"s":       a.S,
		"o":       a.O,
		"tags":    a.Tags,
		"names":   a.Names,
		"del":     a.Del,
	}
}

// struct to map
func (a *Place) Iterate(f func(k PlaceIndexType, v any)) {
	for _, x := range PlaceIndexes() {
		f(x, a.Get(x.String()))
	}
}

// Parse []any to json
func ParseTupleToP(r []any) (a P) {
	for pos, x := range r {
		switch PlaceIndexType(pos) {
		case IndexPlaceID:
			a.Set(FieldPlaceID, x)
		case IndexPlaceUP:
			a.Set(FieldPlaceUP, x)
		case IndexPlaceBoost:
			a.Set(FieldPlaceBoost, x)
		case IndexPlacePeople:
			a.Set(FieldPlacePeople, x)
		case IndexPlaceName:
			a.Set(FieldPlaceName, x)
		case IndexPlaceLat:
			a.Set(FieldPlaceLat, x)
		case IndexPlaceLong:
			a.Set(FieldPlaceLong, x)
		case IndexPlaceHash:
			a.Set(FieldPlaceHash, x)
		case IndexPlaceCountry:
			a.Set(FieldPlaceCountry, x)
		case IndexPlaceState:
			a.Set(FieldPlaceState, x)
		case IndexPlaceZip:
			a.Set(FieldPlaceZip, x)
		case IndexPlaceAlias:
			a.Set(FieldPlaceAlias, x)
		case IndexPlaceTZ:
			a.Set(FieldPlaceTZ, x)
		case IndexPlaceFlag:
			a.Set(FieldPlaceFlag, x)
		case IndexPlaceCur:
			a.Set(FieldPlaceCur, x)
		case IndexPlaceSymbol:
			a.Set(FieldPlaceSymbol, x)
		case IndexPlaceLang:
			a.Set(FieldPlaceLang, x)
		case IndexPlaceS:
			a.Set(FieldPlaceS, x)
		case IndexPlaceO:
			a.Set(FieldPlaceO, x)
		case IndexPlaceTags:
			a.Set(FieldPlaceTags, x)
		case IndexPlaceNames:
			a.Set(FieldPlaceNames, x)
		case IndexPlaceDel:
			a.Set(FieldPlaceDel, x)
		}
	}
	return
}

// NewP create struct
func NewP() P {
	return []byte("{}")
}

// P is a struct
type P []byte

// Set value
func (a *P) Set(k string, v any) *P {
	(*a) = jsons.Set((*a), k, v)
	return a
}

// Get value
func (a *P) Get(k string) jsons.Result {
	return jsons.Get((*a), k)
}

// Get value
func (a *P) DeleteFields(fields ...string) {
	(*a) = jsons.Delete((*a), fields...)
}

// ID set or get value
func (a *P) ID(v ...int) (res int) {
	if v == nil {
		return jsons.Int((*a), FieldPlaceID)
	}
	a.Set(FieldPlaceID, v[0])
	return
}

// UP set or get value
func (a *P) UP(v ...int64) (res int64) {
	if v == nil {
		return jsons.Int64((*a), FieldPlaceUP)
	}
	a.Set(FieldPlaceUP, v[0])
	return
}

// UPTime get value as time
func (a *P) UPTime() (res time.Time) {
	return time.Unix(a.UP(), 0)
}

// Boost set or get value
func (a *P) Boost(v ...int64) (res int64) {
	if v == nil {
		return jsons.Int64((*a), FieldPlaceBoost)
	}
	a.Set(FieldPlaceBoost, v[0])
	return
}

// People set or get value
func (a *P) People(v ...int) (res int) {
	if v == nil {
		return jsons.Int((*a), FieldPlacePeople)
	}
	a.Set(FieldPlacePeople, v[0])
	return
}

// Name set or get value
func (a *P) Name(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceName)
	}
	a.Set(FieldPlaceName, v[0])
	return
}

// Lat set or get value
func (a *P) Lat(v ...float64) (res float64) {
	if v == nil {
		return jsons.Float64((*a), FieldPlaceLat)
	}
	a.Set(FieldPlaceLat, v[0])
	return
}

// Long set or get value
func (a *P) Long(v ...float64) (res float64) {
	if v == nil {
		return jsons.Float64((*a), FieldPlaceLong)
	}
	a.Set(FieldPlaceLong, v[0])
	return
}

// Hash set or get value
func (a *P) Hash(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceHash)
	}
	a.Set(FieldPlaceHash, v[0])
	return
}

// Country set or get value
func (a *P) Country(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceCountry)
	}
	a.Set(FieldPlaceCountry, v[0])
	return
}

// State set or get value
func (a *P) State(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceState)
	}
	a.Set(FieldPlaceState, v[0])
	return
}

// Zip set or get value
func (a *P) Zip(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceZip)
	}
	a.Set(FieldPlaceZip, v[0])
	return
}

// Alias set or get value
func (a *P) Alias(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceAlias)
	}
	a.Set(FieldPlaceAlias, v[0])
	return
}

// TZ set or get value
func (a *P) TZ(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceTZ)
	}
	a.Set(FieldPlaceTZ, v[0])
	return
}

// Flag set or get value
func (a *P) Flag(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceFlag)
	}
	a.Set(FieldPlaceFlag, v[0])
	return
}

// Cur set or get value
func (a *P) Cur(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceCur)
	}
	a.Set(FieldPlaceCur, v[0])
	return
}

// Symbol set or get value
func (a *P) Symbol(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceSymbol)
	}
	a.Set(FieldPlaceSymbol, v[0])
	return
}

// Lang set or get value
func (a *P) Lang(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceLang)
	}
	a.Set(FieldPlaceLang, v[0])
	return
}

// S set or get value
func (a *P) S(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceS)
	}
	a.Set(FieldPlaceS, v[0])
	return
}

// O set or get value
func (a *P) O(v ...string) (res []string) {
	if v == nil {
		return jsons.ArrayString((*a), FieldPlaceO)
	}
	a.Set(FieldPlaceO, v)
	return
}

// OAdd add values
func (a *P) OAdd(v ...string) {
	a.Set(FieldPlaceO, append(a.O(), v...))
}

// OPrepend add values
func (a *P) OPrepend(v ...string) {
	a.Set(FieldPlaceO, append(v, a.O()...))
}

// OAddUnique add unique values only
func (a *P) OAddUnique(v ...string) {
	var list []string
	un := map[string]bool{}
	for _, x := range a.O() {
		if un[x] {
			continue
		}
		list = append(list, x)
		un[x] = true
	}
	for _, x := range v {
		if un[x] {
			continue
		}
		list = append(list, x)
		un[x] = true
	}

	a.O(list...)
}

// ODelete add unique values only
func (a *P) ODelete(v string) {
	var list []string
	for _, x := range a.O() {
		if x == v {
			continue
		}
		list = append(list, x)
	}
	a.O(list...)
}

// Tags set or get value
func (a *P) Tags(v ...string) (res string) {
	if v == nil {
		return jsons.String((*a), FieldPlaceTags)
	}
	a.Set(FieldPlaceTags, v[0])
	return
}

// Names set or get value
func (a *P) Names(v ...map[string]string) (res map[string]string) {
	if v == nil {
		return jsons.MapString((*a), FieldPlaceNames)
	}
	a.Set(FieldPlaceNames, v[0])
	return
}

// NamesAdd add values
func (a *P) NamesAdd(k string, v string) {
	maps := a.Names()
	maps[k] = v
	a.Names(maps)
}

// NamesDelete add unique values only
func (a *P) NamesDelete(k string) {
	maps := a.Names()
	delete(maps, k)
	a.Names(maps)
}

// NamesHas check value
func (a *P) NamesHas(k string) bool {
	maps := a.Names()
	return !reflect.ValueOf(maps[k]).IsZero()
}

// Del set or get value
func (a *P) Del(v ...bool) (res bool) {
	if v == nil {
		return jsons.Bool((*a), FieldPlaceDel)
	}
	a.Set(FieldPlaceDel, v[0])
	return
}
