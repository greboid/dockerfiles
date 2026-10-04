#!/bin/sh
# Hashes pin the geodata snapshot (geonames dumps drift upstream).
mkdir -p geodata && cd geodata
curl -sfLo cities500.zip https://download.geonames.org/export/dump/cities500.zip
echo "a99f1423b13c52d51e281d27fa924c2b11f05a57a96467ff7b8ea437e0b21b38  cities500.zip" | sha256sum -c -
curl -sfLo admin1CodesASCII.txt https://download.geonames.org/export/dump/admin1CodesASCII.txt
echo "1da92a6323a5fec3176f3f743bf4cf4040fd56a876da55e46fbca23c863aa60a  admin1CodesASCII.txt" | sha256sum -c -
curl -sfLo admin2Codes.txt https://download.geonames.org/export/dump/admin2Codes.txt
echo "63a82be925b8704abae683a89e32f40b22f618a93efc4e676e912873f8a3be20  admin2Codes.txt" | sha256sum -c -
curl -sfLo countryInfo.txt https://download.geonames.org/export/dump/countryInfo.txt
echo "93bafc525813f22e4711ff9ed6d626343094ce48c26388dc7c49189b3d7d5512  countryInfo.txt" | sha256sum -c -
curl -sfLo ne_10m_admin_0_countries.geojson https://raw.githubusercontent.com/nvkelso/natural-earth-vector/v5.1.2/geojson/ne_10m_admin_0_countries.geojson
echo "239eec57ac17f100a11e2536cffc56752c318b50ae765b0918ff7aab4ce8f255  ne_10m_admin_0_countries.geojson" | sha256sum -c -
unzip -o cities500.zip
rm cities500.zip
chmod 444 *
date -u -d "@${SOURCE_DATE_EPOCH:-0}" +"%Y-%m-%dT%H:%M:%S+00:00" \
  | tr -d "\n" > geodata-date.txt
cd /home/build
