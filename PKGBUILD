# Maintainer: Amadeus <email>
pkgname=gemhelp
pkgver=r4.f94a0ec
pkgrel=1
pkgdesc="A terminal help command that references Gemini with local man and tldr pages"
arch=('x86_64' 'aarch64')
url="https://github.com/AmadeusDE/gemhelp"
license=('OSL-3.0')
depends=('glibc')
makedepends=('go')
provides=('tldr')
conflicts=('tldr' 'tealdeer' 'tealdeer-git' 'tlrc' 'tlrc-bin' 'nodejs-tldr' 'nodejs-tldr-git')
source=()
sha256sums=()

pkgver() {
	cd "$startdir"
	printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build() {
	cd "$startdir"
	export GOPATH="$srcdir/gopath"
	go build -trimpath -o gemhelp -ldflags="-s -w" ./src
}

package() {
	cd "$startdir"
	install -Dm755 gemhelp "$pkgdir/usr/bin/gemhelp"
	ln -s gemhelp "$pkgdir/usr/bin/tldr"
	ln -s gemhelp "$pkgdir/usr/bin/wiki"
}
