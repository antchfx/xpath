package xpath

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func TestLoadingCache(t *testing.T) {
	c := NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			switch v := key.(type) {
			case int:
				return strconv.Itoa(v), nil
			default:
				return nil, errors.New("invalid type")
			}
		},
		2) // cap = 2
	assertEqual(t, 0, len(c.m))
	v, err := c.get(1)
	assertNoErr(t, err)
	assertEqual(t, "1", v)
	assertEqual(t, 1, len(c.m))

	v, err = c.get(1)
	assertNoErr(t, err)
	assertEqual(t, "1", v)
	assertEqual(t, 1, len(c.m))

	v, err = c.get(2)
	assertNoErr(t, err)
	assertEqual(t, "2", v)
	assertEqual(t, 2, len(c.m))

	// over capacity, m is reset
	v, err = c.get(3)
	assertNoErr(t, err)
	assertEqual(t, "3", v)
	assertEqual(t, 1, len(c.m))

	// Invalid capacity
	assertPanic(t, func() {
		NewLoadingCache(func(key interface{}) (interface{}, error) { return key, nil }, -1)
	})

	// Loading failure
	c = NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			if key.(int)%2 == 0 {
				return key, nil
			} else {
				return nil, fmt.Errorf("artificial error: %d", key.(int))
			}
		}, 0)
	v, err = c.get(12)
	assertNoErr(t, err)
	assertEqual(t, 12, v)
	_, err = c.get(21)
	assertErr(t, err)
	assertEqual(t, "artificial error: 21", err.Error())
}

const (
	benchLoadingCacheRandSeed    = 12345
	benchLoadingCacheConcurrency = 5
	benchLoadingCacheKeyRange    = 2000
	benchLoadingCacheCap         = 1000
)

func BenchmarkLoadingCacheCapped_SingleThread(b *testing.B) {
	rand.Seed(benchLoadingCacheRandSeed)
	c := NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			return key, nil
		}, benchLoadingCacheCap)
	for i := 0; i < b.N; i++ {
		k := rand.Intn(benchLoadingCacheKeyRange)
		v, _ := c.get(k)
		if k != v {
			b.FailNow()
		}
	}
	b.Logf("N=%d, reset=%d", b.N, c.reset)
}

func BenchmarkLoadingCacheCapped_MultiThread(b *testing.B) {
	rand.Seed(benchLoadingCacheRandSeed)
	c := NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			return key, nil
		}, benchLoadingCacheCap)
	wg := sync.WaitGroup{}
	wg.Add(benchLoadingCacheConcurrency)
	for i := 0; i < benchLoadingCacheConcurrency; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				k := rand.Intn(benchLoadingCacheKeyRange)
				v, _ := c.get(k)
				if k != v {
					b.FailNow()
				}
			}
			defer wg.Done()
		}()
	}
	wg.Wait()
	b.Logf("N=%d, concurrency=%d, reset=%d", b.N, benchLoadingCacheConcurrency, c.reset)
}

func BenchmarkLoadingCacheNoCap_SingleThread(b *testing.B) {
	rand.Seed(benchLoadingCacheRandSeed)
	c := NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			return key, nil
		}, 0) // 0 => no cap
	for i := 0; i < b.N; i++ {
		k := rand.Intn(benchLoadingCacheKeyRange)
		v, _ := c.get(k)
		if k != v {
			b.FailNow()
		}
	}
	b.Logf("N=%d, reset=%d", b.N, c.reset)
}

func BenchmarkLoadingCacheNoCap_MultiThread(b *testing.B) {
	rand.Seed(benchLoadingCacheRandSeed)
	c := NewLoadingCache(
		func(key interface{}) (interface{}, error) {
			return key, nil
		}, 0) // 0 => no cap
	wg := sync.WaitGroup{}
	wg.Add(benchLoadingCacheConcurrency)
	for i := 0; i < benchLoadingCacheConcurrency; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				k := rand.Intn(benchLoadingCacheKeyRange)
				v, _ := c.get(k)
				if k != v {
					b.FailNow()
				}
			}
			defer wg.Done()
		}()
	}
	wg.Wait()
	b.Logf("N=%d, concurrency=%d, reset=%d", b.N, benchLoadingCacheConcurrency, c.reset)
}

func TestGetRegexp(t *testing.T) {
	RegexpCache = defaultRegexpCache()
	assertEqual(t, 0, len(RegexpCache.m))
	assertEqual(t, defaultCap, RegexpCache.cap)
	exp, err := getRegexp("^[0-9]{3,5}$")
	assertNoErr(t, err)
	assertTrue(t, exp.MatchString("3141"))
	assertFalse(t, exp.MatchString("3"))
	exp, err = getRegexp("[invalid")
	assertErr(t, err)
	assertEqual(t, "error parsing regexp: missing closing ]: `[invalid`", err.Error())
	assertNil(t, exp)
}
