package ltpool

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
	poolId  int
	payload []byte
}

func (i testPoolItem) GetPoolId() int {
	return i.poolId
}

func (i *testPoolItem) SetPoolId(id int) {
	i.poolId = id
}

func (i *testPoolItem) fill() {
	if i.payload == nil {
		i.payload = make([]byte, 16)
		_, _ = rand.Read(i.payload)
	}
}

func TestPool(t *testing.T) {
	p := NewPool()

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
	if len(p.items) != len(p.free)+1 {
		t.Error("reserved and free items count mismatch")
	}
	if p.statGet != p.statPut {
		t.Error("stat mismatch", p.statGet, "gets vs", p.statPut, "puts")
	}
}

func TestPoolParallel(t *testing.T) {
	p := NewPool()

	for i := 0; i < 100; i++ {
		var wg sync.WaitGroup

		for g := 0; g < 1000; g++ {
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
	if len(p.items) != len(p.free)+1 {
		t.Error("reserved and free items count mismatch")
	}
	if p.statGet != p.statPut {
		t.Error("stat mismatch", p.statGet, "gets vs", p.statPut, "puts")
	}
}

func BenchmarkPool(b *testing.B) {
	p := NewPool()
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
