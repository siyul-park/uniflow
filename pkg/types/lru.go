package types

import (
	"container/list"
	"sync"
	"time"
)

// LRU represents an LRU (Least Recently Used) cache.
type LRU struct {
	capacity int
	entries  map[uint64][]*list.Element
	order    *list.List
	ttl      time.Duration
	mu       sync.Mutex
}

// NewLRU creates a new LRU with the given capacity and TTL.
func NewLRU(capacity int, ttl time.Duration) *LRU {
	return &LRU{
		capacity: capacity,
		entries:  make(map[uint64][]*list.Element),
		order:    list.New(),
		ttl:      ttl,
	}
}

// Load retrieves the value associated with the key.
func (c *LRU) Load(key Value) (Value, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := HashOf(key)
	for _, elem := range c.entries[hash] {
		pair := elem.Value.([3]any)
		if Equal(pair[0].(Value), key) {
			if c.ttl > 0 && time.Now().After(pair[2].(time.Time)) {
				c.remove(elem)
				return nil, false
			}
			c.order.MoveToFront(elem)
			return pair[1].(Value), true
		}
	}
	return nil, false
}

// Store adds or updates a key-value pair in the cache.
func (c *LRU) Store(key, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := HashOf(key)
	for _, elem := range c.entries[hash] {
		pair := elem.Value.([3]any)
		if Equal(pair[0].(Value), key) {
			if c.ttl > 0 {
				elem.Value = [3]any{key, value, time.Now().Add(c.ttl)}
			} else {
				elem.Value = [3]any{key, value, nil}
			}
			c.order.MoveToFront(elem)
			return
		}
	}

	var elem *list.Element
	if c.ttl > 0 {
		elem = c.order.PushFront([3]any{key, value, time.Now().Add(c.ttl)})
	} else {
		elem = c.order.PushFront([3]any{key, value, nil})
	}
	c.entries[hash] = append(c.entries[hash], elem)

	c.evict()
}

// Delete removes a key-value pair from the cache.
func (c *LRU) Delete(key Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := HashOf(key)
	for _, elem := range c.entries[hash] {
		pair := elem.Value.([3]any)
		if Equal(pair[0].(Value), key) {
			c.remove(elem)
			break
		}
	}
}

// Len returns the current number of items in the cache.
func (c *LRU) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.order.Len()
}

// Clear removes all items from the cache.
func (c *LRU) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[uint64][]*list.Element)
	c.order.Init()
}

func (c *LRU) evict() {
	if c.ttl > 0 {
		now := time.Now()
		for elem := c.order.Back(); elem != nil; elem = c.order.Back() {
			pair := elem.Value.([3]any)
			if now.After(pair[2].(time.Time)) {
				c.remove(elem)
			} else {
				break
			}
		}
	}

	for c.capacity > 0 && c.order.Len() > c.capacity {
		elem := c.order.Back()
		if elem == nil {
			return
		}
		c.remove(elem)
	}
}

func (c *LRU) remove(elem *list.Element) {
	pair := elem.Value.([3]any)
	hash := HashOf(pair[0].(Value))

	c.order.Remove(elem)

	for i, e := range c.entries[hash] {
		if e == elem {
			c.entries[hash] = append(c.entries[hash][:i], c.entries[hash][i+1:]...)
			if len(c.entries[hash]) == 0 {
				delete(c.entries, hash)
			}
			break
		}
	}
}
