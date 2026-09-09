// PKGPATH: gno.land/r/demo/crossinpdemo
package crossinpdemo

// gnolang/gno#5786: a package that is not a realm may declare a
// function whose only parameter is a realm. Entering it does not cross:
// the realm reported inside is the caller's.

import "filetests/extern/crossinp"

func main(cur realm) {
	println(crossinp.Sync(cross(cur)))
}

// Output:
// gno.land/r/demo/crossinpdemo
