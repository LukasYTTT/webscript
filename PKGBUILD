# Maintainer: Lukas <lukas@example.com>
pkgname=webscript-git
pkgver=1.0.0
pkgrel=1
pkgdesc="A high-performance reverse proxy and web server language with its own package manager"
arch=('x86_64')
url="https://github.com/yourusername/webscript"
license=('MIT')
depends=('glibc')
makedepends=('go')
provides=('webscript' 'wbs')
conflicts=('webscript')
# For a real AUR package, source would be a git URL. Here we use local files for testing.
source=()
md5sums=()

build() {
  # In a real PKGBUILD, we would cd into the downloaded src dir.
  # cd "$srcdir/$pkgname"
  
  # For local testing, assuming the PKGBUILD is in the project root:
  go build -o wbs .
}

package() {
  # Install the executable to /usr/bin/wbs
  install -Dm755 wbs "$pkgdir/usr/bin/wbs"
  # Optional: symlink webscript to wbs
  ln -s /usr/bin/wbs "$pkgdir/usr/bin/webscript"
}
