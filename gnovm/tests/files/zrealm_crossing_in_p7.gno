// PKGPATH: gno.land/r/demo/crossinpmethod
package crossinpmethod

// The method form, and the one place cross() still differs from a plain
// hand-over: an explicit cross() marks a revive boundary by author
// intent, whether or not a realm changed, so the panic routes past this
// frame's recover. Called as Boom(cur) it recovers normally.

import "filetests/extern/crossinp_adv"

var h = &crossinp_adv.Holder{N: 1}

func main(cur realm) {
	h.Bump(cross(cur))
	println(h.N)
	defer func() {
		println("recovered:", recover())
	}()
	crossinp_adv.Boom(cross(cur))
}

// Output:
// 2

// Error:
// boom
