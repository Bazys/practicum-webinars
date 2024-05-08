package main

import "fmt"

type Comparator[T comparable] interface {
	Compare(a, b T) bool
}

type GenericComparator[T comparable] struct{}

func (gc GenericComparator[T]) Compare(a, b T) bool {
	return a == b
}

func main() {
	intComp := GenericComparator[int]{}
	fmt.Println(intComp.Compare(5, 5)) // Выведет: true

	stringComp := GenericComparator[string]{}
	fmt.Println(stringComp.Compare("hello", "world")) // Выведет: false

	floatComp := GenericComparator[float64]{}
	fmt.Println(floatComp.Compare(3.1415996, 3.1415996)) // Выведет: true
}
