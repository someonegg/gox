// Copyright 2025 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncx

import (
	"sync"
	"sync/atomic"
)

type syncFunc struct {
	f func()

	lock sync.Mutex
	once *callOnce
}

type callOnce struct {
	sync.Once
	enter atomic.Bool
}

func (s *syncFunc) Call() {
	s.lock.Lock()
	once := s.once
	s.lock.Unlock()

	if once.enter.Load() {
		// wait for last call
		once.Do(func() {})

		// get the latest call
		s.lock.Lock()
		once = s.once
		s.lock.Unlock()
	}

	once.Do(func() {
		once.enter.Store(true)
		s.f()

		s.lock.Lock()
		s.once = &callOnce{}
		s.lock.Unlock()
	})
}

// SyncFunc returns a function that invokes f serially. It will also merge calls
// and ensure that there is a complete f() during w().
func SyncFunc(f func()) (w func()) {
	sf := &syncFunc{
		f:    f,
		once: &callOnce{},
	}
	return func() {
		sf.Call()
	}
}
