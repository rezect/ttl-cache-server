package cache

import (
	"fmt"
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
	wg    sync.WaitGroup
}

func CacheNew() *Cache {
	c := new(Cache)
	c.items = make(map[string]item)
	c.done = make(chan struct{})
	c.wg.Add(1)
	go c.cleanupCacheLoop()

	return c
}

func (c *Cache) cleanupCacheLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			c.wg.Done()
			return
		case <-ticker.C:
			c.mu.Lock()
			for cacheKey, cacheItem := range c.items {
				currentExpireDate := cacheItem.expire
				if time.Now().After(currentExpireDate) {
					delete(c.items, cacheKey)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Cache) Get(key string) (any, error) {
	c.mu.RLock()
	currentItem, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	exp := currentItem.expire
	val := currentItem.value

	if time.Now().After(exp) {
		return nil, fmt.Errorf("key has expired")
	}

	return val, nil
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("TTL is not positive")
	}

	expire_time := time.Now().Add(ttl)
	new_item := item{
		value:  value,
		expire: expire_time,
	}

	c.mu.Lock()
	c.items[key] = new_item
	c.mu.Unlock()

	return nil
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	_, ok := c.items[key]

	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("key is not found")
	}

	delete(c.items, key)
	c.mu.Unlock()

	return nil
}

func (c *Cache) Clear() {
	c.mu.Lock()
	for key := range c.items {
		delete(c.items, key)
	}
	c.mu.Unlock()
}

func (c *Cache) Stop() {
	c.mu.Lock()
	for key := range c.items {
		delete(c.items, key)
	}
	c.mu.Unlock()
	close(c.done)
	c.wg.Wait()
}
