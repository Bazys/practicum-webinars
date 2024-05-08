package main

import (
	"fmt"
)

type Integer interface {
	~int | ~int64 | ~int32 | ~int16 | ~int8
}

// Scale returns a copy of s with each element multiplied by c.
// This implementation has a problem, as we will see.
func Scale[E Integer](s []E, c E) []E {
	r := make([]E, len(s))
	for i, v := range s {
		r[i] = v * c
	}
	return r
}

type Point []int32

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p[0], p[1])
}

// ScaleAndPrint doubles a Point and prints it.
func ScaleAndPrint(p Point) {
	r := Scale(p, 2)
	fmt.Println(r)
	//fmt.Println(r.String()) // DOES NOT COMPILE
}

func ScaleFix[S ~[]E, E Integer](s S, c E) S {
	r := make(S, len(s))
	for i, v := range s {
		r[i] = v * c
	}
	return r
}

func ScaleAndPrintFix(p Point) {
	r := ScaleFix(p, 2)
	fmt.Println(r.String()) // DOES NOT COMPILE
}
