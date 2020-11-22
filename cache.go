package xpath

import (
	"regexp"
	"sync"
)

type loadFunc func(key interface{}) (interface{}, error)

const (
	defaultCap = 65536
)

type loadingCache struct {
	sync.RWMutex
	cap   int
	load  loadFunc
	m     map[interface{}]interface{}
	reset int
}

// NewLoadingCache creates a new instance of a loading cache with capacity. Capacity must be >= 0, or
// it will panic. Capacity == 0 means the cache growth is unbounded.
func NewLoadingCache(load loadFunc, capacity int) *loadingCache {
	if capacity < 0 {
		panic("capacity must be >= 0")
	}
	return &loadingCache{cap: capacity, load: load, m: make(map[interface{}]interface{})}
}

func (c *loadingCache) get(key interface{}) (interface{}, error) {
	c.RLock()
	v, found := c.m[key]
	c.RUnlock()
	if found {
		return v, nil
	}
	v, err := c.load(key)
	if err != nil {
		return nil, err
	}
	c.Lock()
	if c.cap > 0 && len(c.m) >= c.cap {
		c.m = map[interface{}]interface{}{key: v}
		c.reset++
	} else {
		c.m[key] = v
	}
	c.Unlock()
	return v, nil
}

var (
	// RegexpCache is a loading cache for string -> *regexp.Regexp mapping. It is exported so that in rare cases
	// client can customize load func and/or capacity.
	RegexpCache = defaultRegexpCache()
)

func defaultRegexpCache() *loadingCache {
	return NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			return regexp.Compile(key.(string))
		}, defaultCap)
}

func getRegexp(pattern string) (*regexp.Regexp, error) {
	exp, err := RegexpCache.get(pattern)
	if err != nil {
		return nil, err
	}
	return exp.(*regexp.Regexp), nil
}
