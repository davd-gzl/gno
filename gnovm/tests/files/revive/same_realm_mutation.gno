// PKGPATH: gno.land/r/test
package test

var counter int

func main() {
	// Test 1: Same-realm mutation should be rolled back
	counter = 0

	ex := revive(func() {
		counter = 42 // This mutation should be rolled back
		panic("test panic")
	})

	if counter != 0 {
		println("FAIL: counter should be 0 but is", counter)
	} else {
		println("PASS: counter correctly rolled back to 0")
	}

	// Test 2: Multiple mutations
	counter = 10

	ex2 := revive(func() {
		counter = 20
		counter = 30
		counter = 40
		panic("test panic 2")
	})

	if counter != 10 {
		println("FAIL: counter should be 10 but is", counter)
	} else {
		println("PASS: counter correctly rolled back to 10")
	}

	println("Exception 1:", ex)
	println("Exception 2:", ex2)
}

// Output:
// PASS: counter correctly rolled back to 0
// PASS: counter correctly rolled back to 10
// Exception 1: test panic
// Exception 2: test panic 2
