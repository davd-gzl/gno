// PKGPATH: gno.land/r/demo/crossinpcur
package crossinpcur

// The migration target. A helper outside /r/ that takes the caller's
// realm is called by handing it over, with no placeholder and no
// cross(), because nothing crosses.

import "filetests/extern/crossinp_adv"

func main(cur realm) {
	println("plain: ", crossinp_adv.Direct(cur))
	println("cross: ", crossinp_adv.Direct(cross(cur)))
}

// Output:
// plain:  gno.land/r/demo/crossinpcur
// cross:  gno.land/r/demo/crossinpcur
