package lazy

import (
	"sync"
	"sync/atomic"
)

type Of[T any] interface {
	Value(ref ...any) T
}

type OfFn[T any] struct {
	done  atomic.Uint32
	m     sync.Mutex
	New   func(ref ...any) T
	value T
}

func MakeFn[T any](newfunc func() T) OfFn[T] {
	return OfFn[T]{New: func(ref ...any) T { return newfunc() }}
}

func MakeRefFn[R, T any](newfunc func(ref R) T) OfFn[T] {
	return OfFn[T]{New: func(ref ...any) T { return newfunc(ref[0].(R)) }}
}

func MakeAnyFn[T any](newfunc func(ref ...any) T) OfFn[T] {
	return OfFn[T]{New: newfunc}
}

func NewFn[T any](newfunc func() T) Of[T] {
	of := MakeFn(newfunc)
	return &of
}

func NewRefFn[R, T any](newfunc func(ref R) T) Of[T] {
	of := MakeRefFn(newfunc)
	return &of
}

func NewAnyFn[T any](newfunc func(ref ...any) T) Of[T] {
	of := MakeAnyFn(newfunc)
	return &of
}

func (this *OfFn[T]) Value(ref ...any) T {
	if this.done.Load() == 0 {
		func() {
			this.m.Lock()
			defer this.m.Unlock()
			if this.done.Load() == 0 {
				this.done.Store(1)
				this.value = this.New(ref...)
			}
		}()
	}
	return this.value
}
