package lbpool

import (
	"crypto/rand"
	"sync"
	"testing"
)

type testPoolNative struct {
	p sync.Pool
}

func (p *testPoolNative) Get() *testPoolItem {
	x := p.p.Get()
	if x != nil {
		if i, ok := x.(*testPoolItem); ok {
			return i
		}
	}
	return &testPoolItem{}
}

func (p *testPoolNative) Put(x *testPoolItem) {
	p.p.Put(x)
}

type testPoolItem struct {
	payload []byte
}

func (i *testPoolItem) Release() {
	// release logic
}

func (i *testPoolItem) fill() {
	if i.payload == nil {
		i.payload = make([]byte, 16)
		_, _ = rand.Read(i.payload)
	}
}

func TestPool(t *testing.T) {
	p := Pool{}

	for i := 0; i < 100; i++ {
		var item *testPoolItem
		x := p.Get()
		if x == nil {
			item = &testPoolItem{}
		} else {
			item = x.(*testPoolItem)
		}
		item.fill()
		p.Put(item)
	}
}

func TestPoolParallel(t *testing.T) {
	p := Pool{}

	for i := 0; i < 100; i++ {
		var wg sync.WaitGroup

		for g := 0; g < 10000; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				var item *testPoolItem
				x := p.Get()
				if x == nil {
					item = &testPoolItem{}
				} else {
					item = x.(*testPoolItem)
				}
				item.fill()
				p.Put(item)
			}()
		}

		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {
	p := Pool{ReleaseFactor: 0.01}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var item *testPoolItem
		x := p.Get()
		if x == nil {
			item = &testPoolItem{}
		} else {
			item = x.(*testPoolItem)
		}
		item.fill()
		p.Put(item)
	}
}

func BenchmarkPoolParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		p := Pool{ReleaseFactor: 0.01}
		for pb.Next() {
			var item *testPoolItem
			x := p.Get()
			if x == nil {
				item = &testPoolItem{}
			} else {
				item = x.(*testPoolItem)
			}
			item.fill()
			p.Put(item)
		}
	})
}

func BenchmarkPoolNative(b *testing.B) {
	p := testPoolNative{}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		item := p.Get()
		if item == nil {
			item = &testPoolItem{}
		}
		item.fill()
		p.Put(item)
	}
}

func BenchmarkPoolNativeParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		p := testPoolNative{}

		for pb.Next() {
			item := p.Get()
			if item == nil {
				item = &testPoolItem{}
			}
			item.fill()
			p.Put(item)
		}
	})
}
