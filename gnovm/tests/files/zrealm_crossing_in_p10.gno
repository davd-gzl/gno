// PKGPATH: gno.land/r/demo/crossinpself
package crossinpself

// A package crossing into its own function is not a cross-package call,
// so it still mints: the callee sees the library package, not the
// caller. This is the idiom p/ test files use to build the frame that
// rlm.Previous() resolves against, and it is deliberately untouched.

import "filetests/extern/crossinp_adv"

func main(cur realm) {
	println("handed over:", crossinp_adv.Direct(cur))
	println("self-cross: ", crossinp_adv.SelfCross(cur))
}

// Output:
// handed over: gno.land/r/demo/crossinpself
// self-cross:  filetests/extern/crossinp_adv
