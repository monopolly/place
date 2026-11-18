package place

import (
	"strings"

	"github.com/monopolly/cast"
	"github.com/monopolly/place/internal/geo"
)

const (
	sourceCityID = int(iota)
	sourceCityName
	sourceCityAsciiname
	sourceCityAlternatenames
	sourceCityLat
	sourceCityLong
	sourceCityFeatureClass
	sourceCityFeatureCode
	sourceCityCountry
	sourceCityCC2
	sourceCityAdmin1
	sourceCityAdmin2
	sourceCityAdmin3
	sourceCityAdmin4
	sourceCityPopulation
	sourceCityElevation
	sourceCityDem
	sourceCityTimezone
	sourceCityDate
)

// cities500.txt
func (a *Engine) parseCitiesLine(line string) (c *Place) {
	c = new(Place)
	var state string
	for pos, value := range strings.Split(line, "\t") {
		value = strings.TrimSpace(value)
		switch pos {
		case sourceCityID:
			c.ID = cast.Int(value)
		case sourceCityAsciiname:
			c.Name = value
		case sourceCityLat:
			c.Lat = cast.Float(value)
		case sourceCityLong:
			c.Long = cast.Float(value)
		case sourceCityCountry:
			c.Country = strings.ToLower(value)
		case sourceCityAdmin1:
			state = value
		case sourceCityPopulation:
			c.People = cast.Int(value)
		case sourceCityTimezone:
			c.TZ = value
		}
	}

	if a.skip(c.ID) {
		return nil
	}

	if a.skipItems[c.ID] != nil {
		return nil
	}

	if c.People < 10 {
		return nil
	}

	switch c.Country {
	case "th":
		c.Name = strings.Replace(c.Name, "Ko ", "Koh ", -1)
	case "us":
		c.State = strings.ToUpper(state)
	case "gb":
		c.Country = "uk"
	}

	// if c.Name == "London" {
	// 	fmt.Println(c.ID, c.Name, c.Country)
	// }

	info := geo.Lookup(c.Country)
	c.Flag = info.Flag
	c.Cur = info.Cur
	c.Symbol = info.Symbol
	c.Lang = info.Lang
	c.Hash = geo.Geohash(c.Lat, c.Long)

	return
}

func (a *Engine) skip(id int) (skip bool) {
	switch id {
	case 1819730, 1819727: //hk
		return true
	default:
		return false
	}
}

/*
	0 5128581
	1 New York City
	2 New York City
	3 Aebura,Bandar Raya New York,Big Apple,Cathair Nua Eabhraic,City of New York,Eabhraig Nuadh,Efrog Newydd,Evrek Nowydh,Ga-no-no,Kanon:no,Kanono,Kanón:no,Kota New York,Lungsod ng New York,N'ju-Jork,NY,NYC,Nea Yorke,Nei Yarrick Schtadt,Neu Amsterdam,Neu Jorck,New Amsterdam,New Orange,New Yorc,New York,New York City,New York Stad,New York borg,New York kenti,New York-borg,New Yorke,New Yorku,Niujorkas,Nju Jork,Njujork,Nouvelle Yorck,Nouvieau York,Nov-Jorko,Nova Amsterda,Nova Iorque,Nova York,Nova-York,Novjorko,Novum Eboracum,Nowy Jork,Nua-Eabhrac,Nueva York,Nujorka,Nyja Jorvik,Nyu York Shehiri,Nòva York,Nýja Jórvík,Thanh pho New York,The Big Apple,Thành phố New York,York Berri,manhattan,n'yuyorka,ni'u iyarka siti,niu yue,niu yue shi,niyuyark nakaram,nkhr niwyxrk,nyuyog,nyuyog si,nyuyoku,nyw ywrq,nywywrk,Ņujorka,Νέα Υόρκη,Њу Јорк,Њујорк,Нью-Ёрк,Нью-Йорк,Ню Йорк,ניו יארק,ניו יורק,ניו־יאָרק,نيويورك,نیویارک شہر,نیویورک,न्यूयॉर्क,নিউ ইয়র্ক সিটি,நியூயார்க் நகரம்,นครนิวยอร์ก,ნიუ-იორკი,ニューヨーク,マンハッタン,紐約市,纽约,뉴욕,뉴욕 시
	4 40.71427
	5 -74.00597
	6	P
	7	PPL
	8	US
	9	NY
	10	8175133
	10
	57
	America/New_York
	2015-02-02 */

/*
	1120473	Wulêswālī Bihsūd	Wuleswali Bihsud	Behsud,Behsūd,Markaz-e Bihsud,Markaz-e Bihsūd,Olusvali Behsowd,Olusvālī Behsowd,Oluswali Behsud,Oluswālī Behsūḏ,Ulusvali-Bikhsud,Wuleswali Bihsud,Wuluswali Bihsud,Wuluswālī Bihsūd,Wulêswālī Bihsūd,mrkz bhswd,wlswaly bhswd,مرکز بهسود,ولسوالی بهسود	34.3436	67.90567	P	PPLA2	AF		27				0		2960	Asia/Kabul	2012-08-04 */

/*
	0 geonameid         : integer id of record in geonames database
	1 name              : name of geographical point (utf8) varchar(200)
	2 asciiname         : name of geographical point in plain ascii characters, varchar(200)
	3 alternatenames    : alternatenames, comma separated, ascii names automatically transliterated, convenience attribute from alternatename table, varchar(10000)
	4 latitude          : latitude in decimal degrees (wgs84)
	5 longitude         : longitude in decimal degrees (wgs84)
	6 feature class     : see http://www.geonames.org/export/codes.html, char(1)
	7 feature code      : see http://www.geonames.org/export/codes.html, varchar(10)
	8 country code      : ISO-3166 2-letter country code, 2 characters
	9 cc2               : alternate country codes, comma separated, ISO-3166 2-letter country code, 200 characters
	10 admin1 code       : fipscode (subject to change to iso code), see exceptions below, see file admin1Codes.txt for display names of this code; varchar(20)
	11 admin2 code       : code for the second administrative division, a county in the US, see file admin2Codes.txt; varchar(80)
	12 admin3 code       : code for third level administrative division, varchar(20)
	13 admin4 code       : code for fourth level administrative division, varchar(20)
	14 population        : bigint (8 byte int)
	15 elevation         : in meters, integer
	16 dem               : digital elevation model, srtm3 or gtopo30, average elevation of 3''x3'' (ca 90mx90m) or 30''x30'' (ca 900mx900m) area in meters, integer. srtm processed by cgiar/ciat.
	17 timezone          : the iana timezone id (see file timeZone.txt) varchar(40)
	18 modification date : date of last modification in yyyy-MM-dd format
*/
