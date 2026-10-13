package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}
type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	mu       sync.Mutex
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	c := &LRUCache[K, V]{
		capacity: max(0, capacity),
		mu:       sync.Mutex{},
		ll:       list.List{},
		items:    make(map[K]*list.Element),
	}
	c.ll.Init()
	return c
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key] // what if other goroutine does c.items[key] = newValue or delete(c.items, key)?
	if !ok || c.capacity == 0 {
		var zero V
		return zero, false
	}

	c.ll.MoveToFront(elem) // what if other goroutine does c.ll.Remove(oldestElem) or delete(c.items, oldestElemKey)?

	ent := elem.Value.(entry[K, V]) // what if c.Set() happened right now?
	return ent.value, true
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.capacity == 0 {
		return
	}

	elem, alreadyExists := c.items[key] // what if elem got deleted just now? panic
	if alreadyExists {                  // what if 2 goroutines try to insert 'x' that's absent? => 2 elems in list, 1 in map
		entCopy := elem.Value.(entry[K, V])
		entCopy.value = value
		elem.Value = entCopy
		return
	}

	cacheIsFull := len(c.items) == c.capacity // 2 Sets on cap-1 size => size = cap+1
	if cacheIsFull {
		oldestElem := c.ll.Back() // what if it just got deleted?
		oldestElemKey := oldestElem.Value.(entry[K, V]).key
		c.ll.Remove(oldestElem)
		delete(c.items, oldestElemKey)
	}

	newEntry := entry[K, V]{key: key, value: value}
	c.ll.PushFront(newEntry)
	c.items[key] = c.ll.Front() // what if front was just changed by other goroutine?
}
