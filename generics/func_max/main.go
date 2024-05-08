package main

import (
	"cmp"
	"fmt"
)

// Max возвращает максимальное значение
// в слайсе любого сравниваемого типа.
func Max[T cmp.Ordered](s []T) T {
	if len(s) == 0 {
		panic("slice is empty")
	}
	m := s[0]
	for i := range s[1:] {
		m = max(m, s[i])
	}
	return m
}

func main() {
	intSlice := []int{1, 3, 5, 7, 9}
	fmt.Println("Maximum integer:", Max(intSlice))

	floatSlice := []float64{2.3, 4.5, 1.1, 3.3}
	fmt.Println("Maximum float:", Max(floatSlice))

	stringSlice := []string{"apple", "orange", "banana", "mango"}
	fmt.Println("Maximum string:", Max(stringSlice))
}
