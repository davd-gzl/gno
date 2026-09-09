// PKGPATH: gno.land/r/demo/crossinpalloc
package crossinpalloc

// An object the helper allocates belongs to the caller's realm, so a
// realm global may keep it.

import "filetests/extern/crossinp"

var kept *crossinp.Box

func main(cur realm) {
	kept = crossinp.New(cross(cur), 7)
	println(kept.N)
}

// Output:
// 7
