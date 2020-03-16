package lockfreeskiplist

import (
	"flag"
	"math/rand"
	"sync/atomic"
	"testing"
)

var n int

func init() {
	flag.IntVar(&n, "n", 10000, "element count")
}

func assert(b *testing.B, cond bool, message string) {
	if cond {
		return
	}
	b.Fatal(message)
}

// Add n elements per goroutines.
func BenchmarkRandomAdd(b *testing.B) {
	var sl = NewLockFreeSkipList(func(value1, value2 interface{}) bool {
		v1, _ := value1.(int)
		v2, _ := value2.(int)
		return v1 < v2
	})
	var count int32
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < n; i++ {
				if sl.Add(rand.Int() % n) {
					atomic.AddInt32(&count, 1)
				}
			}
		}
	})
	assert(b, sl.GetSize() == count, "sl.GetSize() == count")
}

// Remove n elements per goroutines.
func BenchmarkRandomRemove(b *testing.B) {
	var sl = NewLockFreeSkipList(func(value1, value2 interface{}) bool {
		v1, _ := value1.(int)
		v2, _ := value2.(int)
		return v1 < v2
	})
	var count int32
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.StopTimer()
			for i := 0; i < n; i++ {
				if sl.Add(rand.Int() % n) {
					atomic.AddInt32(&count, 1)
				}
			}
			b.StartTimer()
			for i := 0; i < n; i++ {
				if sl.Remove(rand.Int() % n) {
					atomic.AddInt32(&count, -1)
				}
			}
		}
	})
	assert(b, sl.GetSize() == count, "sl.GetSize() == count")
}

// Add and Remove n elements per goroutines.
func BenchmarkRandomAddAndRemoveAndContains(b *testing.B) {
	var sl = NewLockFreeSkipList(func(value1, value2 interface{}) bool {
		v1, _ := value1.(int)
		v2, _ := value2.(int)
		return v1 < v2
	})
	var count int32
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < n; i++ {
				if sl.Add(rand.Int() % n) {
					atomic.AddInt32(&count, 1)
				}
			}
		}
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < n; i++ {
				if sl.Remove(rand.Int() % n) {
					atomic.AddInt32(&count, -1)
				}
			}
		}
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < n; i++ {
				sl.Contains(rand.Int() % n)
			}
		}
	})
	assert(b, sl.GetSize() == count, "sl.GetSize() == count")
}
