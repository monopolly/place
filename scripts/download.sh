#!/bin/sh
# Скачивает дампы geonames для тестов и бенчмарков.
#
#   scripts/download.sh                  # cities15000 + alternateNamesV2
#   scripts/download.sh cities500        # другой набор городов (+ alternateNamesV2)
#   PLACE_DATA=/tmp/place scripts/download.sh
#
# Файлы кладутся в $PLACE_DATA/source (по умолчанию testdata/source) — туда же,
# где их ищет Engine с Volume("testdata"), поэтому повторно он их не качает.
set -eu

cd "$(dirname "$0")/.."

DATA="${PLACE_DATA:-testdata}"
BASE="https://download.geonames.org/export/dump"
CITIES="${1:-cities15000}"

mkdir -p "$DATA/source"

# папка лежит в Dropbox: большие дампы синхронизировать незачем
if command -v xattr >/dev/null 2>&1; then
	xattr -w com.dropbox.ignored 1 "$DATA" 2>/dev/null || true
	xattr -w 'com.apple.fileprovider.ignore#P' 1 "$DATA" 2>/dev/null || true
fi

fetch() {
	name="$1"
	out="$DATA/source/$name.zip"
	if [ -s "$out" ] && unzip -tq "$out" >/dev/null 2>&1; then
		echo "ok       $out"
		return
	fi
	echo "download $BASE/$name.zip"
	curl -fL --retry 3 -C - -o "$out" "$BASE/$name.zip"
	unzip -tq "$out" >/dev/null
	echo "ok       $out"
}

fetch "$CITIES"
fetch alternateNamesV2

ls -lh "$DATA/source"
