# Maintainer: Amadeus <email>
pkgname=gemhelp
pkgver=1.0.0
pkgrel=1
pkgdesc="A terminal help command that references Gemini with local man and tldr pages"
arch=('x86_64' 'aarch64' 'arm64' '386')
url="https://github.com/AmadeusDE/gemhelp"
license=('OSL-3.0')
depends=('glibc')
makedepends=('go')
source=("$pkgname-$pkgver.tar.gz")
sha256sums=('SKIP')

build() {
	cd "$pkgname-$pkgver"
	export GOPATH="$srcdir/gopath"
	go build -o gemhelp -ldflags="-s -w" ./src
}

package() {
	cd "$pkgname-$pkgver"
	install -Dm755 gemhelp "$pkgdir/usr/bin/gemhelp"
}
