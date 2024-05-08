package generics

type Cache[T any] struct {
	storage map[string]T
}

func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		storage: make(map[string]T),
	}
}

func (c *Cache[T]) Set(key string, value T) {
	c.storage[key] = value
}

func (c *Cache[T]) Get(key string) (T, bool) {
	value, found := c.storage[key]
	if !found {
		return value, false
	}
	return value, true
}
