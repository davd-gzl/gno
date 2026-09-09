// PKGPATH: gno.land/r/demo/crossinpkeep
package crossinpkeep

// A package that stashes the realm it was handed keeps it for the rest
// of the transaction, and no longer than that. The leading-realm form
// gives away exactly what the placeholder form already gave away: the
// caller hands its realm over explicitly in both.

import "filetests/extern/crossinp_adv"

func main(cur realm) {
	crossinp_adv.Keep(cur)
	crossinp_adv.KeepPlaceholder(0, cur)
	println("leading:    ", crossinp_adv.Leaked())
	println("placeholder:", crossinp_adv.Leaked2())
}

// Output:
// leading:     gno.land/r/demo/crossinpkeep
// placeholder: gno.land/r/demo/crossinpkeep
