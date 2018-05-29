package authr

import (
	"container/list"
	"regexp"
	"sync"
)

type regexpCacheEntry struct {
	p string
	r *regexp.Regexp
}

type regexpCache interface {
	add(pattern string, r *regexp.Regexp)
	find(pattern string) (*regexp.Regexp, bool)
}

// regexpListCache is a runtime cache for preventing needless regexp.Compile
// operations since it can be expensive in hot areas.
//
// this cache has an LRU eviction policy and uses a double-linked list to
// efficiently shift/remove entries without introducing more CPU overhead.
//
// this cache *significantly* reduces the expense of regexp operators in authr as
// proven by benchmarks (old is noopRegexpCache, new is regexpListCache):
//
// benchmark                                                   old ns/op     new ns/op     delta
// BenchmarkRegexpOperatorSerial/int(33)~*^foo$=>false-8       3755          336           -91.05%
// BenchmarkRegexpOperatorSerial/"foo-one"~*^Foo=>true-8       5803          315           -94.57%
// BenchmarkRegexpOperatorSerial/"bar-two"~^Bar=>false-8       5678          206           -96.37%
// BenchmarkRegexpOperatorParallel/int(33)~*^foo$=>false-8     1428          417           -70.80%
// BenchmarkRegexpOperatorParallel/"foo-one"~*^Foo=>true-8     6516          427           -93.45%
// BenchmarkRegexpOperatorParallel/"bar-two"~^Bar=>false-8     6194          384           -93.80%
// BenchmarkRegexpOperatorThrash-8                             6019          2355          -60.87%
//
// benchmark                                                   old allocs     new allocs     delta
// BenchmarkRegexpOperatorSerial/int(33)~*^foo$=>false-8       52             3              -94.23%
// BenchmarkRegexpOperatorSerial/"foo-one"~*^Foo=>true-8       29             2              -93.10%
// BenchmarkRegexpOperatorSerial/"bar-two"~^Bar=>false-8       28             1              -96.43%
// BenchmarkRegexpOperatorParallel/int(33)~*^foo$=>false-8     52             3              -94.23%
// BenchmarkRegexpOperatorParallel/"foo-one"~*^Foo=>true-8     29             2              -93.10%
// BenchmarkRegexpOperatorParallel/"bar-two"~^Bar=>false-8     28             1              -96.43%
// BenchmarkRegexpOperatorThrash-8                             36             14             -61.11%
//
// benchmark                                                   old bytes     new bytes     delta
// BenchmarkRegexpOperatorSerial/int(33)~*^foo$=>false-8       3297          32            -99.03%
// BenchmarkRegexpOperatorSerial/"foo-one"~*^Foo=>true-8       39016         24            -99.94%
// BenchmarkRegexpOperatorSerial/"bar-two"~^Bar=>false-8       39008         16            -99.96%
// BenchmarkRegexpOperatorParallel/int(33)~*^foo$=>false-8     3299          32            -99.03%
// BenchmarkRegexpOperatorParallel/"foo-one"~*^Foo=>true-8     39016         24            -99.94%
// BenchmarkRegexpOperatorParallel/"bar-two"~^Bar=>false-8     39008         16            -99.96%
// BenchmarkRegexpOperatorThrash-8                             27071         9086          -66.44%
type regexpListCache struct {
	sync.Mutex
	// capacity
	c int
	// current length
	s int
	l *list.List
}

func newRegexpListCache(capacity int) *regexpListCache {
	if capacity < 0 {
		panic("negative regexp cache")
	}
	return &regexpListCache{
		c: capacity,
		s: 0,
		l: list.New().Init(),
	}
}

func (r *regexpListCache) add(pattern string, _r *regexp.Regexp) {
	r.Lock()
	defer r.Unlock()
	if r.s > r.c {
		panic("regexpListCache overflow")
	}
	if r.s == r.c {
		r.l.Remove(r.l.Back())
		r.s--
	}
	r.l.PushFront(&regexpCacheEntry{p: pattern, r: _r})
	r.s++
}

func (r *regexpListCache) find(pattern string) (*regexp.Regexp, bool) {
	r.Lock()
	defer r.Unlock()
	var e *list.Element = r.l.Front()
	for e != nil {
		if e.Value.(*regexpCacheEntry).p == pattern {
			r.l.MoveToFront(e)
			return e.Value.(*regexpCacheEntry).r, true
		}
		e = e.Next()
	}
	return nil, false
}

type noopRegexpCache struct{}

func (n *noopRegexpCache) add(_ string, _ *regexp.Regexp) {}
func (n *noopRegexpCache) find(_ string) (*regexp.Regexp, bool) {
	return nil, false
}
