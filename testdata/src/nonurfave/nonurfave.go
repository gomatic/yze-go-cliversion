package main

type point struct{ x, y int }

func main() {
	// a main package that builds no urfave command — nothing to enforce.
	_ = point{x: 1, y: 2}
	_ = []string{"x"}
	_ = struct{ A int }{A: 1}
}
