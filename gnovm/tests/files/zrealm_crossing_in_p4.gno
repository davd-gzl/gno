// PKGPATH: gno.land/r/demo/crossinpadv
package crossinpadv

// The leading-realm form and the placeholder form it replaces must be
// indistinguishable from the caller.

import "filetests/extern/crossinp_adv"

func main(cur realm) {
	println("leading:    ", crossinp_adv.Direct(cur))
	println("placeholder:", crossinp_adv.Forward(0, cur))
}

// Output:
// leading:     gno.land/r/demo/crossinpadv
// placeholder: gno.land/r/demo/crossinpadv
