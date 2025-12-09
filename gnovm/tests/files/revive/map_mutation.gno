package main

var m = make(map[string]int)

func main() {
	// Test 1: Map mutation should be rolled back
	m["initial"] = 100
	println("Before revive:", m["initial"])

	revive(func() {
		m["key"] = 42
		m["initial"] = 200
		println("Inside revive:", m["key"], m["initial"])
		panic("test panic")
	})

	println("After revive:", m["key"], m["initial"])

	// Test 2: Nested map operations
	revive(func() {
		m["nested"] = 999
		panic("nested panic")
	})

	println("After nested:", m["nested"])
}

// Output:
// Before revive: 100
// Inside revive: 42 200
// After revive: 0 100
// After nested: 0
