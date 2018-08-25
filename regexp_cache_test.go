package authr

import (
	"regexp"
	"runtime"
	"testing"
)

func BenchmarkListCacheAddSerial(b *testing.B) {
	c := newRegexpListCache(5)
	b.ReportAllocs()
	var _r *regexp.Regexp
	for i := 0; i < b.N; i++ {
		c.add("a", _r)
	}
}

func BenchmarkListCacheAddParallel(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	b.SetParallelism(runtime.NumCPU())
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.add("a", _r)
		}
	})
}

func BenchmarkListCacheFindMissParallel(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	c.add("a", _r)
	c.add("b", _r)
	c.add("c", _r)
	c.add("d", _r)
	c.add("e", _r)
	b.SetParallelism(runtime.NumCPU())
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.find("f")
		}
	})
}

func BenchmarkListCacheFindMissSerial(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	c.add("a", _r)
	c.add("b", _r)
	c.add("c", _r)
	c.add("d", _r)
	c.add("e", _r)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.find("f")
	}
}

func BenchmarkListCacheFindHitStartParallel(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	c.add("a", _r)
	c.add("b", _r)
	b.SetParallelism(runtime.NumCPU())
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.find("b")
		}
	})
}

func BenchmarkListCacheFindHitStartSerial(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	c.add("a", _r)
	c.add("b", _r)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.find("b")
	}
}

func BenchmarkListCacheFindHitEndParallel(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	c.add("a", _r)
	c.add("b", _r)
	c.add("c", _r)
	c.add("d", _r)
	c.add("e", _r)
	b.SetParallelism(runtime.NumCPU())
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.find("e")
		}
	})
}

func BenchmarkListCacheFindHitEndSerial(b *testing.B) {
	c := newRegexpListCache(5)
	var _r *regexp.Regexp
	c.add("a", _r)
	c.add("b", _r)
	c.add("c", _r)
	c.add("d", _r)
	c.add("e", _r)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.find("e")
	}
}

func TestListCache(t *testing.T) {
	t.Parallel()
	t.Run("should store and be able to find", func(t *testing.T) {
		c := newRegexpListCache(5)
		var _r *regexp.Regexp
		c.add("ozncowoldu", _r)
		r, ok := c.find("ozncowoldu")
		if !ok {
			t.Fatalf("unexpected cache miss")
			return
		}
		if r != _r {
			t.Fatalf("cache returned the wrong pattern? %p != %p", _r, r)
			return
		}
	})
	t.Run("should miss if not able to find pattern", func(t *testing.T) {
		c := newRegexpListCache(5)
		r, ok := c.find("sckvccisjm")
		if ok {
			t.Fatalf("unexpected cache hit")
			return
		}
		if r != nil {
			t.Fatalf("unexpected *regexp.Regexp returned: %+v", r)
			return
		}
	})
	t.Run("should start overflowing and removing stuff", func(t *testing.T) {
		c := newRegexpListCache(5)
		var _r *regexp.Regexp = &regexp.Regexp{}
		c.add("mjepcahoxe", _r)
		c.add("qpafzozhjf", _r)
		c.add("wbdporssdz", _r)

		// fetch 'mjepcahoxe', this should move to the front
		rr, ok := c.find("mjepcahoxe")
		if !ok {
			t.Fatalf("unexpected cache miss for 'mjepcahoxe'")
			return
		}
		if rr == nil {
			t.Fatalf("unexpected nil *regexp.Regexp for 'mjepcahoxe'")
			return
		}

		c.add("znzqyktuuw", _r)
		c.add("isuteoxatj", _r)
		c.add("pkzbgrkdff", _r)
		c.add("wncwhcpjsh", _r)
	})
}
