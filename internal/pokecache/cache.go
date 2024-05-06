package pokecache

import (
	"sync"
	"time"
)

var mu = &sync.RWMutex{}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries map[string]cacheEntry
	reap    time.Duration
}

func NewCache(d time.Duration) Cache {
	var C Cache = Cache{
		entries: make(map[string]cacheEntry, 0),
	}
	ticker := time.NewTicker(d)

	go func() {
		for range ticker.C {
			C.reapLoop()
		}
	}()

	return C
}

func (c Cache) reapLoop() {
	for key, entry := range c.entries {
		if time.Now().Sub(entry.createdAt) > c.reap {
			delete(c.entries, key)
		}
	}
}

func (c Cache) Add(key string, val []byte) {
	mu.Lock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	mu.Unlock()
}

func (c Cache) Get(key string) ([]byte, bool) {
	mu.RLock()
	defer mu.RUnlock()
	val, ok := c.entries[key]
	if !ok {
		return nil, ok
	}
	return val.val, ok
}
