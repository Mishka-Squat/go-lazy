package lazy

import (
	"sync"
	"sync/atomic"
)

type Of[T any] interface {
	Value(ref ...any) T
}

type OfM[T any] interface {
	SetValue(value T)
}

type OfFn[T any] struct {
	done  atomic.Uint32
	m     sync.Mutex
	New   func(ref ...any) T
	value T
}

type OfFnCached[T any] struct {
	OfFn[T]
	key any
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

func (this *OfFn[T]) mutex_condition_fn(cond func() bool, fn func(), els ...func()) {
	if cond() {
		func() {
			this.m.Lock()
			defer this.m.Unlock()
			if cond() {
				fn()
			}
		}()
	} else {
		for _, fn := range els {
			fn()
		}
	}
}

func (this *OfFn[T]) Value(ref ...any) T {
	this.mutex_condition_fn(func() bool { return this.done.Load() == 0 },
		func() {
			this.done.Store(1)
			this.value = this.New(ref...)
		})
	return this.value
}

func (this *OfFnCached[T]) Value(ref ...any) T {
	this.mutex_condition_fn(func() bool { return this.done.Load() == 0 },
		func() {
			this.done.Store(1)
			this.value = this.New(ref...)
		}, func() {
			if len(ref) > 0 {
				this.mutex_condition_fn(func() bool { return this.key != ref[0] },
					func() {

						this.done.Store(0)
					})
			}
		})
	return this.value
}

func (this *OfFn[T]) SetValue(value T) {
	this.done.Store(1)
	this.value = value
}
