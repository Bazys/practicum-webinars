// Пример generic-метода (Go 1.27+).
package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Tree[T any] struct {
	Value    T
	Children []*Tree[T]
}

func NewTree[T any](value T, children ...*Tree[T]) *Tree[T] {
	return &Tree[T]{Value: value, Children: children}
}

// Map — generic-метод: кроме параметра ресивера T объявляет свой параметр U.
// До Go 1.27 это можно было написать только свободной функцией
// Map[T, U any](t *Tree[T], f func(T) U) *Tree[U].
func (t *Tree[T]) Map[U any](f func(T) U) *Tree[U] {
	if t == nil {
		return nil
	}
	r := &Tree[U]{Value: f(t.Value)}
	for _, c := range t.Children {
		r.Children = append(r.Children, c.Map(f))
	}
	return r
}

// Fold сворачивает дерево в значение произвольного типа A.
func (t *Tree[T]) Fold[A any](acc A, f func(A, T) A) A {
	if t == nil {
		return acc
	}
	acc = f(acc, t.Value)
	for _, c := range t.Children {
		acc = c.Fold(acc, f)
	}
	return acc
}

// Ограничение: метод не может сузить T ресивера (any → comparable),
// поэтому поиск по значению по-прежнему остаётся свободной функцией.
func Contains[T comparable](t *Tree[T], v T) bool {
	return t.Fold(false, func(found bool, x T) bool { return found || x == v })
}

// Ограничение: generic-метод не реализует метод интерфейса,
// и в самом интерфейсе метод не может иметь параметров типа.
//
// type Mapper interface {
//     Map[U any](func(int) U) any // ошибка компиляции
// }

func main() {
	tree := NewTree(
		1,
		NewTree(2, NewTree(4)),
		NewTree(3),
	)

	// U выводится из аргумента: *Tree[int] → *Tree[string]
	strs := tree.Map(strconv.Itoa)
	fmt.Printf("%T\n", strs) // *main.Tree[string]

	// Явная инстанциация метода тоже возможна
	floats := tree.Map[float64](func(v int) float64 { return float64(v) / 2 })
	fmt.Println(floats.Children[0].Value) // 1

	sum := tree.Fold(0, func(acc, v int) int { return acc + v })
	fmt.Println("sum:", sum) // sum: 10

	joined := strs.Fold("", func(acc, v string) string {
		return strings.TrimPrefix(acc+","+v, ",")
	})
	fmt.Println("values:", joined) // values: 1,2,4,3

	fmt.Println(Contains(tree, 4), Contains(tree, 5)) // true false
}
