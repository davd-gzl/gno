// PKGPATH: gno.land/r/demo/crossinpwrite
package crossinpwrite

// The helper writes a field of an object the caller owns. It can,
// because it runs with the caller's realm rather than one of its own.

import "filetests/extern/crossinp"

var b = &crossinp.Box{N: 1}

func main(cur realm) {
	crossinp.Mutate(cross(cur), b)
	println(b.N)
}

// Output:
// 42
