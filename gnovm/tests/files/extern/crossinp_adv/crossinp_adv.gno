package crossinp_adv

// Adversarial surface for a crossing function declared outside a realm.
// Each entry point is one thing a package could try to do with a realm
// the caller handed it.

var stash realm

// Keep tries to hold the caller's realm past the call.
func Keep(cur realm) {
	stash = cur
}

// Leaked reports what Keep held.
func Leaked() string {
	if stash == nil {
		return "<nil>"
	}
	return stash.PkgPath()
}

// Direct reports the realm it was handed. Forward is the same body
// reached through the placeholder form; the two must agree.
func Direct(cur realm) string {
	return cur.PkgPath()
}

func Forward(_ int, rlm realm) string {
	return rlm.PkgPath()
}

// SelfCross is the other case, kept separate because it is not a
// cross-package call: a package crossing into its own function still
// mints, so the callee sees this package rather than the caller. p/ test
// files rely on it to build the frame rlm.Previous() resolves against.
func SelfCross(cur realm) string {
	return inner(cross(cur))
}

func inner(cur realm) string {
	return cur.PkgPath()
}

// Boom panics from inside a crossing function the caller entered with
// cross(). Panic routing has to find the caller's revive frame.
func Boom(cur realm) {
	panic("boom")
}

type Holder struct {
	N int
}

// Bump is the method form: a receiver does not push the realm off
// index 0, so this was unwritable too.
func (h *Holder) Bump(cur realm) {
	h.N++
}

// BoomPlaceholder is Boom written the old way. A panic out of either
// must reach the caller the same way.
func BoomPlaceholder(_ int, rlm realm) {
	panic("boom")
}

var stash2 realm

// KeepPlaceholder is Keep written the old way, for comparison.
func KeepPlaceholder(_ int, rlm realm) {
	stash2 = rlm
}

func Leaked2() string {
	if stash2 == nil {
		return "<nil>"
	}
	return stash2.PkgPath()
}
