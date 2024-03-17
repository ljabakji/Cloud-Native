package cache

import "errors"

type Cacher[K comparable, V any] interface {
	Get(key K) (value V, err error)
	Put(key K, value V) (err error)
}

// Concrete LRU cache
type lruCache[K comparable, V any] struct {
	size      int
	remaining int
	cache     map[K]V
	queue     []K
}

// Constructor
func NewCacher[K comparable, V any](size int) Cacher[K, V] {
	return &lruCache[K, V]{size: size, remaining: size, cache: make(map[K]V), queue: make([]K, 0)}
}

// Get method retrieves a value for a given key and updates the queue
func (c *lruCache[K, V]) Get(key K) (value V, err error) {
	val, ok := c.cache[key]
	if !ok {
		// Key does not exist, return an error
		var zeroVal V // Needed to return a zero value of V
		return zeroVal, errors.New("key not found")
	}

	// Move the key to the tail of the queue to mark as recently used
	c.deleteFromQueue(key)         // Remove existing occurrences of the key
	c.queue = append(c.queue, key) // Append key to the tail

	// Return the found value
	return val, nil
}

// Put method adds a new key-value pair to the cache or updates an existing key
func (c *lruCache[K, V]) Put(key K, value V) (err error) {
	_, exists := c.cache[key]
	if exists || len(c.cache) < c.size {
		// If key exists or there is space, update/add the key-value pair
		if !exists {
			c.remaining-- // Decrement remaining space if adding a new key
		}
		c.cache[key] = value
	} else {
		// Evict the least recently used item (at the head of the queue)
		oldestKey := c.queue[0]
		delete(c.cache, oldestKey)   // Remove from cache
		c.deleteFromQueue(oldestKey) // Clean up the queue
		c.cache[key] = value         // Add the new key-value pair
	}

	c.deleteFromQueue(key)         // Ensure the key is only once in the queue
	c.queue = append(c.queue, key) // Append key to the tail to mark as recently used

	return nil
}

// deleteFromQueue removes all occurrences of a key from the queue
func (c *lruCache[K, V]) deleteFromQueue(key K) {
	newQueue := make([]K, 0)
	for _, qKey := range c.queue {
		if qKey != key {
			newQueue = append(newQueue, qKey) // Keep key if it's not the one to delete
		}
	}
	c.queue = newQueue // Update the queue without the deleted key
}
