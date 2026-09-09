package mytoken

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestTransfer(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	uassert.Equal(t, TotalSupply(), int64(10_000_000_000))
	uassert.Equal(t, BalanceOf(owner), int64(10_000_000_000))

	testing.SetOriginCaller(owner)
	Transfer(cross(cur), alice, 500)

	uassert.Equal(t, BalanceOf(alice), int64(500))
	uassert.Equal(t, BalanceOf(owner), int64(9_999_999_500))
}
