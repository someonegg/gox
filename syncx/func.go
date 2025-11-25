// Copyright 2025 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncx

import "sync"

type SyncFunc struct {
	fn func(interface{}) error

	cond  *sync.Cond
	once  *callOnce
	doing bool
}

type callOnce struct {
	sync.Once
	err error
}

func (s *SyncFunc) Call(arg interface{}) error {
	s.cond.L.Lock()
	if s.doing {
		s.cond.Wait()
	}
	s.doing = true
	once := s.once
	s.cond.L.Unlock()

	once.Do(func() {
		once.err = s.fn(arg)

		s.cond.L.Lock()
		s.doing = false
		s.once = &callOnce{}
		s.cond.Broadcast()
		s.cond.L.Unlock()
	})

	return once.err
}

func NewSyncFunc(fn func(interface{}) error) *SyncFunc {
	return &SyncFunc{
		fn:   fn,
		cond: sync.NewCond(&sync.Mutex{}),
		once: &callOnce{},
	}
}
