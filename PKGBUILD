# Maintainer: Lukas <lukas@example.com>
pkgname=webscript-git
pkgver=1.0.0
pkgrel=1
pkgdesc="A high-performance reverse proxy and web server language"
arch=('x86_64')
url="https://github.com/LukasYTTT/webscript"
license=('MIT')
makedepends=('go' 'git')
provides=('webscript' 'wbs')
conflicts=('webscript')
source=("git+https://github.com/LukasYTTT/webscript.git")
md5sums=('SKIP')

pkgver() {
  cd "$srcdir/${pkgname%-git}"
  printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build() {
  cd "$srcdir/${pkgname%-git}"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -extldflags ${LDFLAGS}"
  go build -o wbs .
}

package() {
  cd "$srcdir/${pkgname%-git}"
  install -Dm755 wbs "$pkgdir/usr/bin/wbs"
  ln -s /usr/bin/wbs "$pkgdir/usr/bin/webscript"
}
