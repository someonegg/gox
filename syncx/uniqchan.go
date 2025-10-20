// Copyright 2025 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncx

import (
	"context"
	"sync"
)

// UniqChan represents a channel that filters duplicates.
type UniqChan[K comparable, V any] struct {
	l sync.Mutex
	c chan K
	m map[K]V
}

func NewUniqChan[K comparable, V any](cap int) *UniqChan[K, V] {
	if cap <= 0 {
		panic("cap must > 0")
	}
	return &UniqChan[K, V]{
		c: make(chan K, cap),
		m: make(map[K]V, cap),
	}
}

func (uc *UniqChan[K, V]) Send(k K, v V, updateV bool) {
	uc.SendContext(context.Background(), k, v, updateV)
}

func (uc *UniqChan[K, V]) SendContext(ctx context.Context, k K, v V, updateV bool) (err error) {
	uc.l.Lock()
	_, found := uc.m[k]
	if !found || updateV {
		uc.m[k] = v
	}
	uc.l.Unlock()

	if found {
		return
	}

	select {
	case uc.c <- k:
	case <-ctx.Done():
		err = ctx.Err()

		uc.l.Lock()
		delete(uc.m, k)
		uc.l.Unlock()
	}

	return
}

func (uc *UniqChan[K, V]) TrySend(k K, v V, updateV bool) (full bool) {
	uc.l.Lock()
	defer uc.l.Unlock()

	_, found := uc.m[k]
	if !found || updateV {
		uc.m[k] = v
	}

	if found {
		return
	}

	select {
	case uc.c <- k:
	default:
		full = true
		delete(uc.m, k)
	}

	return
}

func (uc *UniqChan[K, V]) Recv() (k K, v V) {
	k, v, _ = uc.RecvContext(context.Background())
	return
}

func (uc *UniqChan[K, V]) RecvContext(ctx context.Context) (k K, v V, err error) {
	select {
	case k = <-uc.c:
	case <-ctx.Done():
		err = ctx.Err()
		return
	}

	uc.l.Lock()
	v = uc.m[k]
	delete(uc.m, k)
	uc.l.Unlock()

	return
}
