package main

type Point struct {
	X, Y int
}

type Container struct {
	P     Point
	Value int
}

var p = Point{X: 10, Y: 20}
var c = Container{P: Point{X: 1, Y: 2}, Value: 100}

func main() {
	// Test 1: Simple struct mutation should be rolled back
	println("Before revive - p:", p.X, p.Y)

	revive(func() {
		p.X = 99
		p.Y = 88
		println("Inside revive - p:", p.X, p.Y)
		panic("test panic")
	})

	println("After revive - p:", p.X, p.Y)

	// Test 2: Nested struct mutation should be rolled back
	println("Before revive - c:", c.P.X, c.P.Y, c.Value)

	revive(func() {
		c.P.X = 500
		c.P.Y = 600
		c.Value = 999
		println("Inside revive - c:", c.P.X, c.P.Y, c.Value)
		panic("nested panic")
	})

	println("After revive - c:", c.P.X, c.P.Y, c.Value)
}

// Output:
// Before revive - p: 10 20
// Inside revive - p: 99 88
// After revive - p: 10 20
// Before revive - c: 1 2 100
// Inside revive - c: 500 600 999
// After revive - c: 1 2 100
