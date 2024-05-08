package main

import "fmt"

type Tree[T any] struct {
	Value    T
	Children []*Tree[T]
}

// NewTree создает новый узел дерева с заданным значением.
func NewTree[T any](value T) *Tree[T] {
	return &Tree[T]{Value: value}
}

// AddChild добавляет дочерний узел к узлу дерева.
func (t *Tree[T]) AddChild(child *Tree[T]) {
	t.Children = append(t.Children, child)
}

// FindInTree рекурсивно ищет значение в дереве и возвращает true, если значение найдено.
func FindInTree[T comparable](t *Tree[T], value T) bool {
	if t == nil {
		return false
	}
	if t.Value == value {
		return true
	}
	for _, child := range t.Children {
		if FindInTree(child, value) {
			return true
		}
	}
	return false
}

func main() {
	// Создание корня дерева и дочерних узлов
	root := NewTree("root")
	child1 := NewTree("child1")
	child2 := NewTree("child2")
	child11 := NewTree("child1_1")

	// Строим дерево
	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(child11)

	// Поиск значения в дереве
	fmt.Println("Is 'child1_1' in tree?", FindInTree(root, "child1_1")) // ожидаем true
	fmt.Println("Is 'unknown' in tree?", FindInTree(root, "unknown"))   // ожидаем false
}
