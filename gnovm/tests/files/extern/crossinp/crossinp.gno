package crossinp

// A crossing function declared outside a realm. Two things had to change
// for this to mean anything: the ban on declaring it, and the assumption
// that entering it switches realms. It does not. Only a realm crosses,
// so `cur` here is the caller's realm and writes land in the caller's
// storage, which is what the discarded leading parameter was for:
// func Sync(_ int, rlm realm) string.

func Sync(cur realm) string {
	return cur.PkgPath()
}

type Box struct {
	N int
}

// Mutate writes through a pointer the caller owns.
func Mutate(cur realm, b *Box) {
	b.N = 42
}

// New allocates and hands the object back to the caller. It belongs to
// the caller's realm, so it survives finalization.
func New(cur realm, n int) *Box {
	return &Box{N: n}
}
