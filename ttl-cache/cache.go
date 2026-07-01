package cache

import (
	"sync"
	"time"
)

type item struct {
	value  any
	expire time.Time
}

type Cache struct {
	items map[string]item
	mu    sync.RWMutex
	done  chan struct{}
}

func CacheNew() *Cache {
	c := new(Cache)
	c.items = make(map[string]item)
	c.done = make(chan struct{})
	go c.cleenupCacheLoop()

	return c
}

func (c *Cache) cleenupCacheLoop() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			for cacheKey, cacheItem := range c.items {
				currentExpireDate := cacheItem.expire
				if time.Now().After(currentExpireDate) {
					c.mu.Unlock()
					delete(c.items, cacheKey)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	currentItem, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	exp := currentItem.expire
	val := currentItem.value

	if time.Now().After(exp) {
		c.mu.Lock()
		currentItem, ok = c.items[key]
		if !ok {
			c.mu.Unlock()
			return nil, false
		}

		exp = currentItem.expire
		val = currentItem.value

		if time.Now().After(exp) {
			delete(c.items, key)
			c.mu.Unlock()
			return nil, false
		}
		c.mu.Unlock()
	}

	return val, true
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	expire_time := time.Now().Add(ttl)
	new_item := item{
		value:  value,
		expire: expire_time,
	}

	c.mu.Lock()
	c.items[key] = new_item
	c.mu.Unlock()
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	_, ok := c.items[key]

	if !ok {
		c.mu.Unlock()
		return
	}

	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]item)
	c.mu.Unlock()
}

func (c *Cache) Stop() {
	c.mu.Lock()
	c.items = make(map[string]item)
	close(c.done)
	c.mu.Unlock()
}
