package adder

func Sum[T Adder[T]](a, b T) T {
	return a.Add(a, b)
}

type Adder[T any] interface {
	Add(a, b T) T
}

type Container[T Adder[T]] struct {
	Value T
}

func (c *Container[T]) Add(a, b *Container[T]) T {
	return c.Value.Add(a.Value, b.Value)
}

func (c *Container[T]) Sum(a, b *Container[T]) T {
	return c.Add(a, b)
}
