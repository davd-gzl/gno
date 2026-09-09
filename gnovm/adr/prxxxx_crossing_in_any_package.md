# A package may take the caller's realm as its first parameter

## Context

[gnolang/gno#5786](https://github.com/gnolang/gno/issues/5786): a `/p/`
package cannot declare a helper whose only, or first, parameter is a
`realm`. A leading `realm` makes the function crossing, and crossing was
allowed only in a realm. The workaround is a parameter that holds the
first slot and carries nothing:

```gno
func inviteMembers(_ int, rlm realm, boardID boards.ID, invites ...Invite)
inviteMembers(0, cur, boardID, invites...)
```

The restriction was written for a design where packages never crossed.
Under interrealm v2 they do, so the rule refuses a signature that has a
meaning.

## Decision

Two halves, and the second is what makes the first worth anything.

**Any package may declare a crossing function.** The ban lived in three
places, with its carve-outs in a fourth: `crossingAllowed` and its two
call sites, `isRealm`, the `!m.Package.IsRealm()` gate in
`doOpEnterCrossing`, and `crossingFromTestFile`, which existed only to
punch a `*_test.gno` hole in that gate. All removed.

**Only a realm crosses.** Entering such a function does not switch
realms. The callee runs with the caller's realm and receives the
caller's `cur`, which is what the placeholder form was for:

```go
func crossKeepsCallerRealm(caller *PackageValue, callee *FuncValue) bool {
	if caller == nil || callee == nil {
		return false
	}
	if callee.IsClosure {
		return false
	}
	if caller.PkgPath == callee.PkgPath {
		return false
	}
	return isImmutableLibraryPath(callee.PkgPath)
}
```

Each clause is load-bearing, and each was forced by a failing test
rather than chosen up front.

`isImmutableLibraryPath` is already the predicate
[`fillPackage`](https://github.com/gnolang/gno/blob/abcd20dad/gnovm/pkg/gnolang/store.go#L599-L611)
uses to hand a package a frozen realm. `/p/` and the stdlibs own no
storage, so there is nothing to switch to. `/r/`, `/e/` run scripts and
`*_test` overlays are excluded and mint as before.

A call that stays inside one package is excluded because
`f(cross(cur))` on a sibling is the self-cross idiom, which mints a
fresh cur for the same realm. `/p/` test files write
`func(cur realm){...}(cross(cur))` to build the frame `rlm.Previous()`
resolves against.

A function literal is excluded because it is the declaring code's own
body, reached through whatever helper was handed it.
[`uassert.AbortsContains`](https://github.com/gnolang/gno/blob/abcd20dad/examples/gno.land/p/nt/uassert/v0/uassert.gno#L135)
takes a `func(realm)` from a test and cross-calls it back across a
package boundary.

`f(cur)` is accepted alongside `f(cross(cur))`, so migrating a call site
is a deletion of `0, ` rather than a rewrite. Writing `cross` for a
callee that does not cross would claim a boundary that never opens.

## Alternatives considered

**Move the realm last.**
[#6033](https://github.com/gnolang/gno/pull/6033). Clears the placeholder
from 116 signatures and leaves 68 with nowhere to put a trailing realm.

**Read the leading parameter's name**, `cur` crossing and `rlm` not, with
`FuncType.TypeID` carrying the distinction so the two shapes are
different types. It works, and it makes a parameter name change a type,
which Go does nowhere else.

**Delete the ban and nothing else.** Then a `/p/` crossing function
genuinely crosses into a realm that owns nothing: writing a caller-owned
field raises `cannot directly modify readonly tainted object`, and
returning an object allocated in the frame raises `unexpected unreal
object` from
[`toRefValue`](https://github.com/gnolang/gno/blob/abcd20dad/gnovm/pkg/gnolang/realm.go#L2059-L2061)
during finalization, because a `/p/` realm is never persisted.

## Consequences

A cross-call that keeps the caller's realm is no longer a finalization
boundary, so `maybeFinalize` skips it. Writing the caller's realm back
mid-frame is what the third filetest below catches. It stays a boundary
for panic routing, which keys on the explicit `cross()` marker by author
intent and not on whether a realm changed.

No new authority. The only way into such a function is for the caller to
hand its realm over, and `cross(rlm)` still runs its IsCurrent-strict
check. That is the same grant the placeholder form already made, pinned
by three filetests that run both spellings side by side: the realm each
reports, what each leaks when the callee stashes it, and how a panic out
of each reaches the caller.

Eleven filetests cover the shape in a non-realm package, the method
form, `f(cur)` against `f(cross(cur))`, the finalizer under an
unfinalized caller object, and the self-cross that still mints.
