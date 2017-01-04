package bahamut

import (
	"sync"
	"time"
)

type memoryCache struct {
	data       map[string]*cacheItem
	lock       *sync.Mutex
	expiration time.Duration
}

// NewMemoryCache returns a new generic cache.
func NewMemoryCache() Cacher {

	return &memoryCache{
		data:       map[string]*cacheItem{},
		lock:       &sync.Mutex{},
		expiration: -1,
	}
}

func (c *memoryCache) SetDefaultExpiration(exp time.Duration) {

	c.expiration = exp
}

func (c *memoryCache) Get(id string) interface{} {

	c.lock.Lock()
	item, ok := c.data[id]
	c.lock.Unlock()

	if !ok {
		return nil
	}

	return item.data
}

func (c *memoryCache) Set(id string, item interface{}) {

	c.SetWithExpiration(id, item, c.expiration)
}

func (c *memoryCache) SetWithExpiration(id string, item interface{}, exp time.Duration) {

	var timer *time.Timer
	if exp != -1 {
		timer = time.AfterFunc(exp, func() { c.Del(id) })
	}

	ci := &cacheItem{
		identifier: id,
		data:       item,
		timestamp:  time.Now(),
		timer:      timer,
	}

	c.lock.Lock()
	if item, ok := c.data[id]; ok && item.timer != nil {
		item.timer.Stop()
	}
	c.data[id] = ci
	c.lock.Unlock()
}

func (c *memoryCache) Del(id string) {

	c.lock.Lock()
	if item, ok := c.data[id]; ok && item.timer != nil {
		item.timer.Stop()
	}
	delete(c.data, id)
	c.lock.Unlock()
}

func (c *memoryCache) Exists(id string) bool {

	c.lock.Lock()
	_, ok := c.data[id]
	c.lock.Unlock()

	return ok
}

func (c *memoryCache) All() map[string]interface{} {

	out := map[string]interface{}{}

	c.lock.Lock()
	for k, i := range c.data {
		out[k] = i.data
	}
	c.lock.Unlock()

	return out
}
