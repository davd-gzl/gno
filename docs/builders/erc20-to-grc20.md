# Porting an ERC20 to Gno, line by line

Three pieces of an ERC20 move when it becomes a Gno realm. The balances mapping
becomes a [`grc20.Token`](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/demo/tokens/grc20/token.gno),
`msg.sender` becomes `cur.Previous().Address()`, and `onlyOwner` becomes an
[`ownable.Ownable`](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/nt/ownable/v0/ownable.gno).
The rest is Go.

The result is one file of about eighty lines. It keeps the ERC20 function
names, so an indexer written against `Transfer` and `Approve` still reads it.

## Prerequisites

Complete [Getting started](./getting-started.md). This guide assumes you can
write Go and have deployed one realm.

## The contract we start from

A minimal ERC20, with the owner holding the whole initial supply:

```solidity
contract MyToken {
    string public name = "My Token";
    string public symbol = "MTK";
    uint8 public decimals = 4;
    uint256 public totalSupply;
    address public owner;

    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    modifier onlyOwner() {
        require(msg.sender == owner, "not owner");
        _;
    }

    constructor(uint256 initialSupply) {
        owner = msg.sender;
        _mint(msg.sender, initialSupply);
    }

    function transfer(address to, uint256 value) external returns (bool) { ... }
    function approve(address spender, uint256 value) external returns (bool) { ... }
    function transferFrom(address from, address to, uint256 value) external returns (bool) { ... }
    function mint(address to, uint256 value) external onlyOwner { ... }
}
```

## Where do the two mappings go?

Into one call. `grc20.NewToken` returns a `*grc20.Token` and a
`*grc20.PrivateLedger`, and the balances and allowances live inside them:

```go
Token, privateLedger = grc20.NewToken("My Token", "MTK", 4, 0, cur)
```

The two handles are the access control. `Token` reads and hands out tellers.
`privateLedger` mints, burns and moves any balance without asking, so it stays
unexported and no caller outside the realm can reach it. Solidity gets the same
split from `internal` on `_mint` and `_transfer`.

The fourth argument is a token id. A realm minting one token passes `0`, since
nothing of its own can collide with it. A realm minting several draws each id
from one persistent
[`seqid.ID`](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/nt/seqid/v0/seqid.gno).

## Where did `msg.sender` go?

Into the first parameter. A function that writes state takes a leading
`cur realm`, which makes it crossing, and `cur.Previous()` is the realm or the
user that called in:

```go
func Mint(cur realm, to address, amount int64) {
	Ownable.AssertOwnedBy(cur.Previous().Address())
	checkErr(privateLedger.Mint(to, amount))
}
```

Solidity reads a global. Gno passes the caller as an argument the runtime fills
in, so a function that never declares `cur` can never learn who called it and
can never write state. [Interrealm](../resources/gno-interrealm.md) has the
full rule.

For the ERC20 functions themselves the token does this for you.
`privateLedger.CallerTeller()` returns a teller that resolves the caller on
every write, so `Transfer` is a one-line forward:

```go
func Transfer(cur realm, to address, amount int64) {
	checkErr(userTeller.Transfer(0, cur, to, amount))
}
```

## Where does the constructor go?

Into `init`, which runs once when the realm is added to the chain:

```go
func init(cur realm) {
	Token, privateLedger = grc20.NewToken("My Token", "MTK", 4, 0, cur)
	Ownable = ownable.NewWithAddress(owner)
	userTeller = privateLedger.CallerTeller()

	privateLedger.Mint(owner, 1_000_000*10_000)
}
```

Nothing is passed in at deploy time, so the initial supply and the owner are
constants in the file rather than constructor arguments. Name the owner address
explicitly: under `gno test` there is no deploy transaction, so a realm that
reads its owner from `cur.Previous()` in `init` cannot be tested.

## What replaces `onlyOwner`?

`Ownable`, one line at the top of the function. Gno has no modifiers, so the
check is an ordinary call:

```go
Ownable.AssertOwnedBy(cur.Previous().Address())
```

It panics when the caller is not the owner, which aborts the transaction and
rolls back every write, the same as a failed `require`.

## What replaces `require` and `revert`?

The ledger returns an `error` and the realm decides. `checkErr` panics on a
non-nil error, and a panic reverts the whole transaction:

```go
func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
```

Returning the error instead is a valid choice, and it is the one a realm
calling yours will want.

## Where do the events go?

Nowhere. `grc20` emits them for you. Its write paths call
[`chain.Emit`](https://github.com/gnolang/gno/blob/master/gnovm/stdlibs/chain/emit_event.gno)
with `Transfer` and `Approval`, the same names an ERC20 indexer already knows.

## What has no Solidity equivalent?

`Render`, which gives the token a page:

```go
func Render(path string) string {
	switch {
	case path == "":
		return Token.RenderHome()
	...
	}
}
```

A realm deployed at `gno.land/r/example/mytoken` serves that markdown at that
URL, so the token ships with a balance explorer and needs no front end.

## The whole realm

[embedmd]:# (../_assets/erc20-to-grc20/mytoken.gno go)
```go
package mytoken

import (
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	Token         *grc20.Token
	Ownable       *ownable.Ownable
	privateLedger *grc20.PrivateLedger
	userTeller    grc20.Teller
)

// owner holds the whole initial supply. Replace it with your own address
// before deploying.
const owner = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

// init runs once, when the realm is added to the chain. It replaces the
// Solidity constructor.
func init(cur realm) {
	// This realm creates one token, so the id 0 cannot collide.
	Token, privateLedger = grc20.NewToken("My Token", "MTK", 4, 0, cur)
	Ownable = ownable.NewWithAddress(owner)
	userTeller = privateLedger.CallerTeller()

	privateLedger.Mint(owner, 1_000_000*10_000)
}

func TotalSupply() int64 { return Token.TotalSupply() }

func BalanceOf(owner address) int64 { return Token.BalanceOf(owner) }

func Allowance(owner, spender address) int64 { return Token.Allowance(owner, spender) }

// Transfer moves amount from the caller to to.
func Transfer(cur realm, to address, amount int64) {
	checkErr(userTeller.Transfer(0, cur, to, amount))
}

// Approve lets spender move amount out of the caller's balance.
func Approve(cur realm, spender address, amount int64) {
	checkErr(userTeller.Approve(0, cur, spender, amount))
}

// TransferFrom spends the caller's allowance on from.
func TransferFrom(cur realm, from, to address, amount int64) {
	checkErr(userTeller.TransferFrom(0, cur, from, to, amount))
}

// Mint creates amount and credits to. Only the owner can call it.
func Mint(cur realm, to address, amount int64) {
	Ownable.AssertOwnedBy(cur.Previous().Address())
	checkErr(privateLedger.Mint(to, amount))
}

// Burn destroys amount held by from. Only the owner can call it.
func Burn(cur realm, from address, amount int64) {
	Ownable.AssertOwnedBy(cur.Previous().Address())
	checkErr(privateLedger.Burn(from, amount))
}

// Render gives the token a page at gno.land/r/<namespace>/mytoken.
func Render(path string) string {
	parts := strings.Split(path, "/")

	switch {
	case path == "":
		return Token.RenderHome()
	case len(parts) == 2 && parts[0] == "balance":
		return ufmt.Sprintf("%d\n", Token.BalanceOf(address(parts[1])))
	default:
		return "404\n"
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
```

## Testing it

Tests are Go tests. `testing.SetOriginCaller` picks who is calling, and
`cross(cur)` calls a crossing function the way a transaction would:

[embedmd]:# (../_assets/erc20-to-grc20/mytoken_test.gno go)
```go
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
```

Run them with `gno test .` in the realm directory.

## Deploying it

```sh
gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-namespace>/mytoken" \
  -pkgdir . \
  -gas-fee 1000000ugnot -gas-wanted 20000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 \
  <your-key>
```

## Three differences that will bite you

Amounts are `int64`, not `uint256`. The ceiling is about 9.2 quintillion base
units, so a token with 18 decimals holds at most 9.2 whole units. Pick the
decimals to fit the supply, not out of habit.

Decimals stop at 18, the
[`MaxDecimals`](https://github.com/gnolang/gno/blob/master/examples/gno.land/p/demo/tokens/grc20/types.gno)
constant, and `NewToken` panics past it.

There is no `payable`, no fallback and no `receive`. A function marked nothing
in particular can be sent GNOT, and the realm reads and moves it through a
[`chain/banker`](https://github.com/gnolang/gno/blob/master/gnovm/stdlibs/chain/banker/banker.gno)
rather than through a keyword on the signature.

## Next steps

- [Effective Gno](../resources/effective-gno.md) for the storage rules a token
  hits first.
- [Tutorial: `minisocial` dApp](./tutorial-minisocial.md) for the full local
  development loop with `gnodev`.
