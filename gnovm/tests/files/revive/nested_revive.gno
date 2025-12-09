package main

var counter int

func main() {
	counter = 0
	println("Initial:", counter)
	
	// Outer revive
	revive(func() {
		counter = 10
		println("Outer revive start:", counter)
		
		// Inner revive - should roll back to 10
		revive(func() {
			counter = 20
			println("Inner revive:", counter)
			panic("inner panic")
		})
		
		println("After inner revive:", counter)
		
		// Modify again before outer panic
		counter = 15
		println("Before outer panic:", counter)
		panic("outer panic")
	})
	
	println("After outer revive:", counter)
}

// Output:
// Initial: 0
// Outer revive start: 10
// Inner revive: 20
// After inner revive: 10
// Before outer panic: 15
// After outer revive: 0
