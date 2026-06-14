# Maintainer: Lukas <lukas@example.com>
pkgname=webscript-git
pkgver=r7.747d0d2
pkgrel=1
pkgdesc="A high-performance reverse proxy and web server language"
arch=('x86_64')
url="https://github.com/LukasYTTT/webscript"
license=('MIT')
depends=('glibc')
makedepends=('git' 'go')
provides=('webscript' 'wbs')
conflicts=('webscript')
source=('git+https://github.com/LukasYTTT/webscript.git')
sha256sums=('SKIP')

pkgver() {
  cd webscript
  printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short=7 HEAD)"
}

build() {
  cd webscript
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  go build -trimpath -buildmode=pie -ldflags "-extldflags ${LDFLAGS}" -o wbs .
}

package() {
  cd webscript
  install -Dm755 wbs "$pkgdir/usr/bin/wbs"
  ln -s /usr/bin/wbs "$pkgdir/usr/bin/webscript"
}
