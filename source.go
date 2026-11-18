package place

import "path/filepath"

type source int

const (
	city500 = source(iota)
	city1000
	city5000
	city15000
	alternateNamesV2
)

func (a source) link() string {
	switch a {
	case city1000:
		return "https://download.geonames.org/export/dump/cities1000.zip"
	case city5000:
		return "https://download.geonames.org/export/dump/cities5000.zip"
	case city15000:
		return "https://download.geonames.org/export/dump/cities15000.zip"
	case alternateNamesV2:
		return "https://download.geonames.org/export/dump/alternateNamesV2.zip"
	default:
		return "https://download.geonames.org/export/dump/cities500.zip"
	}
}

func (a source) zipName() string {
	switch a {
	case city1000:
		return filepath.Join(volume, dirSource, "cities1000.zip")
	case city5000:
		return filepath.Join(volume, dirSource, "cities5000.zip")
	case city15000:
		return filepath.Join(volume, dirSource, "cities15000.zip")
	case alternateNamesV2:
		return filepath.Join(volume, dirSource, "alternateNamesV2.zip")
	default:
		return filepath.Join(volume, dirSource, "cities500.zip")
	}
}

func (a source) name() string {
	switch a {
	case city1000:
		return "cities1000"
	case city5000:
		return "cities5000"
	case city15000:
		return "cities15000"
	case alternateNamesV2:
		return "alternateNamesV2"
	default:
		return "cities500"
	}
}

func (a source) filename() string {
	switch a {
	case city1000:
		return filepath.Join(volume, dirSource, "cities1000.txt")
	case city5000:
		return filepath.Join(volume, dirSource, "cities5000.txt")
	case city15000:
		return filepath.Join(volume, dirSource, "cities15000.txt")
	case alternateNamesV2:
		return filepath.Join(volume, dirSource, "alternateNamesV2.txt")
	default:
		return filepath.Join(volume, dirSource, "cities500.txt")
	}
}
func (a source) textname() string {
	switch a {
	case city1000:
		return "cities1000.txt"
	case city5000:
		return "cities5000.txt"
	case city15000:
		return "cities15000.txt"
	case alternateNamesV2:
		return "alternateNamesV2.txt"
	default:
		return "cities500.txt"
	}
}
