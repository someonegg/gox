// Copyright 2025 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncx_test

import (
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/someonegg/gox/syncx"
)

func TestSyncFunc(t *testing.T) {
	n := 0
	w := syncx.SyncFunc(func() {
		time.Sleep(50 * time.Millisecond)
		n++
		t.Log(n)
	})

	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.IntN(200)) * time.Millisecond)
			w()
		}()
	}

	wg.Wait()

	if n > 8 {
		t.Fatal("too many f()")
	}
}
