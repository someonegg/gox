// Copyright 2025 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncx_test

import (
	"testing"

	"github.com/someonegg/gox/syncx"
)

func TestUniqChan(t *testing.T) {
	uc := syncx.NewUniqChan[string, string](3)
	uc.Send("a", "1", true)
	uc.Send("b", "2", false)
	uc.Send("c", "3", true)
	uc.Send("c", "3.1", false)

	k, v := uc.Recv()
	if k != "a" || v != "1" {
		t.Fatalf("got %s %s; want a 1", k, v)
	}
	k, v = uc.Recv()
	if k != "b" || v != "2" {
		t.Fatalf("got %s %s; want b 2", k, v)
	}
	k, v = uc.Recv()
	if k != "c" || v != "3" {
		t.Fatalf("got %s %s; want c 3", k, v)
	}

	uc.Send("a", "1", false)
	uc.Send("b", "2", true)
	uc.Send("c", "3", false)
	uc.Send("c", "3.2", true)

	k, v = uc.Recv()
	if k != "a" || v != "1" {
		t.Fatalf("got %s %s; want a 1", k, v)
	}
	k, v = uc.Recv()
	if k != "b" || v != "2" {
		t.Fatalf("got %s %s; want b 2", k, v)
	}
	k, v = uc.Recv()
	if k != "c" || v != "3.2" {
		t.Fatalf("got %s %s; want c 3.2", k, v)
	}

	uc.Send("a", "1", false)
	uc.Send("b", "2", false)
	uc.Send("c", "3", false)

	full := uc.TrySend("d", "4", false)
	if !full {
		t.Fatalf("want full")
	}
	full = uc.TrySend("c", "3.3", false)
	if full {
		t.Fatalf("want no full")
	}
	full = uc.TrySend("c", "3.3", true)
	if full {
		t.Fatalf("want no full")
	}

	k, v = uc.Recv()
	if k != "a" || v != "1" {
		t.Fatalf("got %s %s; want a 1", k, v)
	}
	k, v = uc.Recv()
	if k != "b" || v != "2" {
		t.Fatalf("got %s %s; want b 2", k, v)
	}
	k, v = uc.Recv()
	if k != "c" || v != "3.3" {
		t.Fatalf("got %s %s; want c 3.3", k, v)
	}
}
