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
	m map[K]ucelem[V]
}

type ucelem[V any] struct {
	v    V
	sent DoneChan
	fail DoneChan
}

func NewUniqChan[K comparable, V any](cap int) *UniqChan[K, V] {
	if cap <= 0 {
		panic("cap must > 0")
	}
	return &UniqChan[K, V]{
		c: make(chan K, cap),
		m: make(map[K]ucelem[V], cap),
	}
}

func (uc *UniqChan[K, V]) Send(k K, v V, updateV bool) {
	uc.SendContext(context.Background(), k, v, updateV)
}

func (uc *UniqChan[K, V]) SendContext(ctx context.Context, k K, v V, updateV bool) error {
retry:
	uc.l.Lock()

	e, found := uc.m[k]
	if !found {
		e.v = v
		e.sent = NewDoneChan()
		e.fail = NewDoneChan()
		uc.m[k] = e
	} else {
		if updateV {
			e.v = v
			uc.m[k] = e
		}
	}

	uc.l.Unlock()

	if found {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-e.sent:
			return nil
		case <-e.fail:
			updateV = false
			goto retry
		}
	}

	select {
	case <-ctx.Done():
		uc.l.Lock()
		delete(uc.m, k)
		uc.l.Unlock()
		e.fail.SetDone()
		return ctx.Err()
	case uc.c <- k:
		e.sent.SetDone()
		return nil
	}
}

func (uc *UniqChan[K, V]) TrySend(k K, v V, updateV bool) (sent bool) {
	uc.l.Lock()
	defer uc.l.Unlock()

	e, found := uc.m[k]
	if !found {
		e.v = v
		e.sent = NewDoneChan()
		e.fail = NewDoneChan()
		uc.m[k] = e
	} else {
		if updateV {
			e.v = v
			uc.m[k] = e
		}
	}

	if found {
		select {
		case <-e.sent:
			return true
		default:
			return false
		}
	}

	select {
	case uc.c <- k:
		e.sent.SetDone()
		return true
	default:
		delete(uc.m, k)
		e.fail.SetDone()
		return false
	}
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
	v = uc.m[k].v
	delete(uc.m, k)
	uc.l.Unlock()

	return
}
