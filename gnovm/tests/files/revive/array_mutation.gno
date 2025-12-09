package main

var arr = [3]int{1, 2, 3}
var slice = []int{10, 20, 30}

func main() {
	// Test 1: Array mutation should be rolled back
	println("Before revive - arr:", arr[0], arr[1], arr[2])

	revive(func() {
		arr[0] = 100
		arr[1] = 200
		println("Inside revive - arr:", arr[0], arr[1], arr[2])
		panic("test panic")
	})

	println("After revive - arr:", arr[0], arr[1], arr[2])

	// Test 2: Slice mutation should be rolled back
	println("Before revive - slice:", slice[0], slice[1], slice[2])

	revive(func() {
		slice[0] = 1000
		slice[2] = 3000
		println("Inside revive - slice:", slice[0], slice[1], slice[2])
		panic("slice panic")
	})

	println("After revive - slice:", slice[0], slice[1], slice[2])
}

// Output:
// Before revive - arr: 1 2 3
// Inside revive - arr: 100 200 3
// After revive - arr: 1 2 3
// Before revive - slice: 10 20 30
// Inside revive - slice: 1000 20 3000
// After revive - slice: 10 20 30
