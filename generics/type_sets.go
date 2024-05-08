package generics

import "fmt"

type Numeric interface {
	int | float64
}

type BaseNumeric interface {
	~int | ~float64
}

func Sum[T BaseNumeric](a, b T) T {
	return a + b
}

type MyInt int

func Summarise() {
	var a, b MyInt = 5, 6
	fmt.Println(Sum(a, b))
}
