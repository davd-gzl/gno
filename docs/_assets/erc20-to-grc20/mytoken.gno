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
