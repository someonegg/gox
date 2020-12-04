// Copyright 2020 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pool

import (
	"container/list"
	"sync"
	"time"
)

type Pool interface {
	Get() interface{}
	Put(x interface{})
}

// NewStandardPool returns a standard sync.Pool.
func NewStandardPool(newFunc func() interface{}) Pool {
	return &sync.Pool{
		New: newFunc,
	}
}

type hugeObjectPool struct {
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	New func() interface{}

	cycle time.Duration

	mu sync.Mutex
	xs *list.List

	activeC chan struct{}
}

// NewHugeObjectPool returns a special pool for huge objects that are
// hundreds of megabytes or bigger.
//
// Using standard pool to store huge objects will have side effects: they
// will use more memory than expected, and even have the risk of OOM.
//
// HugeObjectPool will be more accurate. Its GC is also smoother, evict
// at most one object per gcCycle when it is not zero.
func NewHugeObjectPool(newFunc func() interface{}, gcCycle time.Duration) Pool {
	p := &hugeObjectPool{
		New:     newFunc,
		cycle:   gcCycle,
		xs:      list.New(),
		activeC: make(chan struct{}, 1),
	}
	if p.cycle > 0 {
		go p.gc()
	}
	return p
}

func (p *hugeObjectPool) Get() interface{} {
	var x interface{}

	p.mu.Lock()
	e := p.xs.Front()
	if e != nil {
		x = e.Value
		p.xs.Remove(e)
	}
	p.mu.Unlock()

	if x == nil && p.New != nil {
		x = p.New()
	}

	return x
}

func (p *hugeObjectPool) Put(x interface{}) {
	p.mu.Lock()
	p.xs.PushBack(x)
	p.mu.Unlock()

	select {
	case p.activeC <- struct{}{}:
	default:
	}
}

func (p *hugeObjectPool) gc() {
	t := time.NewTimer(p.cycle)
	if !t.Stop() {
		<-t.C
	}

	for {
		select {
		case <-p.activeC:
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
			t.Reset(p.cycle)
		case <-t.C:
			p.mu.Lock()
			e := p.xs.Front()
			if e != nil {
				p.xs.Remove(e)
			}
			l := p.xs.Len()
			p.mu.Unlock()
			if l > 0 {
				t.Reset(p.cycle)
			}
		}
	}
}
