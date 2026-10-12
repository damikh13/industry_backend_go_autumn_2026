package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	cache := &LRUCache[K, V]{
		capacity: max(0, capacity),
		ll:       list.List{},
		items:    make(map[K]*list.Element),
	}
	cache.ll.Init()
	return cache
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	elem, ok := c.items[key]
	if !ok || c.capacity == 0 {
		var zero V
		return zero, false
	}

	c.ll.MoveToFront(elem)

	ent := elem.Value.(entry[K, V])
	return ent.value, true
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}

	elem, alreadyExists := c.items[key]
	if alreadyExists {
		entryCopy := elem.Value.(entry[K, V])
		entryCopy.value = value
		elem.Value = entryCopy
		return
	}

	cacheIsFull := c.capacity == len(c.items)
	if cacheIsFull {
		oldestElem := c.ll.Back()
		oldestElemKey := oldestElem.Value.(entry[K, V]).key
		c.ll.Remove(oldestElem)
		delete(c.items, oldestElemKey)
	}

	newEntry := entry[K, V]{key: key, value: value}
	c.ll.PushFront(newEntry)
	c.items[key] = c.ll.Front()
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
