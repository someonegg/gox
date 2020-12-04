// Copyright 2020 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pool_test

import (
	"testing"
	"time"

	"github.com/someonegg/gox/pool"
)

func TestHugeObjectPool(t *testing.T) {
	p1 := pool.NewHugeObjectPool(nil, 0)

	p1.Put("a")
	p1.Put("b")
	p1.Put("c")

	if g := p1.Get(); g != "a" {
		t.Fatalf("got %#v; want a", g)
	}
	if g := p1.Get(); g != "b" {
		t.Fatalf("got %#v; want b", g)
	}
	if g := p1.Get(); g != "c" {
		t.Fatalf("got %#v; want c", g)
	}
	if g := p1.Get(); g != nil {
		t.Fatalf("got %#v; want nil", g)
	}

	p2 := pool.NewHugeObjectPool(nil, 2*time.Millisecond)
	p2.Put("a")
	p2.Put("b")
	p2.Put("c")

	if g := p2.Get(); g != "a" {
		t.Fatalf("got %#v; want a", g)
	}

	time.Sleep(3 * time.Millisecond)
	if g := p2.Get(); g != "c" {
		t.Fatalf("got %#v; want c", g)
	}

	if g := p2.Get(); g != nil {
		t.Fatalf("got %#v; want nil", g)
	}
}
