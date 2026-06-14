#!/bin/bash
set -e

VERSION="6.0.7"
ARCH="amd64"
PKG_NAME="webscript_${VERSION}_${ARCH}"

echo "Bauen des Go-Binaries..."
CGO_ENABLED=0 go build -trimpath -o wbs .

echo "Erstellen der Ordnerstruktur..."
mkdir -p ${PKG_NAME}/DEBIAN
mkdir -p ${PKG_NAME}/usr/bin

echo "Kopieren der Dateien..."
cp wbs ${PKG_NAME}/usr/bin/wbs
cd ${PKG_NAME}/usr/bin/
ln -sf wbs webscript
cd ../../../

echo "Erstellen der control-Datei..."
cat <<EOF > ${PKG_NAME}/DEBIAN/control
Package: webscript
Version: ${VERSION}
Architecture: ${ARCH}
Maintainer: Lukas <lukas@example.com>
Description: A high-performance reverse proxy and web server language
EOF

echo "Bauen des .deb Pakets..."
dpkg-deb --build ${PKG_NAME}

echo "Aufräumen..."
rm -rf ${PKG_NAME}

echo "Fertig! Das Paket ${PKG_NAME}.deb wurde erstellt."
