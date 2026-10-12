package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	capacity = max(capacity, 0)
	return &Cache[K, V]{capacity: capacity, items: make(map[K]V, capacity)}
}
func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	val, ok := c.items[k]
	return val, ok
}
func (c *Cache[K, V]) Set(k K, v V) bool {
	_, alreadyExists := c.items[k]
	cacheFull := len(c.items) == c.capacity

	if !alreadyExists && cacheFull {
		return false
	}

	c.items[k] = v
	return true
}
