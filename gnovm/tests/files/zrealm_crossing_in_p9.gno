// PKGPATH: gno.land/r/demo/crossinppanic
package crossinppanic

// A panic out of the leading-realm form must reach the caller exactly as
// it does out of the placeholder form it replaces.

import "filetests/extern/crossinp_adv"

func try(label string, f func()) {
	defer func() {
		println(label, "recovered:", recover())
	}()
	f()
}

func main(cur realm) {
	try("placeholder", func() { crossinp_adv.BoomPlaceholder(0, cur) })
	try("leading    ", func() { crossinp_adv.Boom(cur) })
}

// Output:
// placeholder recovered: boom
// leading     recovered: boom
