package bahamut

import "time"

// A Cacher is the interface caching struct have to implement
type Cacher interface {
	SetDefaultExpiration(exp time.Duration)
	Set(id string, item interface{})
	SetWithExpiration(id string, item interface{}, exp time.Duration)
	Get(id string) interface{}
	Del(id string)
	Exists(id string) bool
	All() map[string]interface{}
}

type cacheItem struct {
	timestamp  time.Time
	identifier string
	data       interface{}
	timer      *time.Timer
}
