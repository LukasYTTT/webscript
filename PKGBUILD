# Maintainer: Lukas <lukas@example.com>
pkgname=webscript-git
pkgver=1.0.0
pkgrel=1
pkgdesc="A high-performance reverse proxy and web server language"
arch=('x86_64')
url="https://github.com/LukasYTTT/webscript"
license=('MIT')
makedepends=('go')
provides=('webscript' 'wbs')
conflicts=('webscript')
source=()

build() {
  cd "$startdir"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -extldflags ${LDFLAGS}"
  go build -o wbs .
}

package() {
  cd "$startdir"
  install -Dm755 wbs "$pkgdir/usr/bin/wbs"
  ln -s /usr/bin/wbs "$pkgdir/usr/bin/webscript"
}
