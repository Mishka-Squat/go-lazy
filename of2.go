package lazy

import (
	"sync"
	"sync/atomic"
)

type Of2[T1, T2 any] interface {
	Value(ref ...any) (T1, T2)
}

type Of2M[T1, T2 any] interface {
	SetValue(v1 T1, v2 T2)
}

type Of2Fn[T1, T2 any] struct {
	done atomic.Uint32
	m    sync.Mutex
	New  func(ref ...any) (T1, T2)
	v1   T1
	v2   T2
}

type Of2FnCached[T1, T2 any] struct {
	Of2Fn[T1, T2]
	key any
}

func Make2Fn[T1, T2 any](newfunc func() (T1, T2)) Of2Fn[T1, T2] {
	return Of2Fn[T1, T2]{New: func(ref ...any) (T1, T2) { return newfunc() }}
}

func Make2RefFn[R, T1, T2 any](newfunc func(ref R) (T1, T2)) Of2Fn[T1, T2] {
	return Of2Fn[T1, T2]{New: func(ref ...any) (T1, T2) { return newfunc(ref[0].(R)) }}
}

func Make2AnyFn[T1, T2 any](newfunc func(ref ...any) (T1, T2)) Of2Fn[T1, T2] {
	return Of2Fn[T1, T2]{New: newfunc}
}

func New2Fn[T1, T2 any](newfunc func() (T1, T2)) Of2[T1, T2] {
	of2 := Make2Fn(newfunc)
	return &of2
}

func New2RefFn[R, T1, T2 any](newfunc func(ref R) (T1, T2)) Of2[T1, T2] {
	of2 := Make2RefFn(newfunc)
	return &of2
}

func New2AnyFn[T1, T2 any](newfunc func(ref ...any) (T1, T2)) Of2[T1, T2] {
	of2 := Make2AnyFn(newfunc)
	return &of2
}

func (this *Of2Fn[T1, T2]) mutex_condition_fn(cond func() bool, fn func(), els ...func()) {
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

func (this *Of2Fn[T1, T2]) Value(ref ...any) (T1, T2) {
	this.mutex_condition_fn(func() bool { return this.done.Load() == 0 },
		func() {
			this.done.Store(1)
			this.v1, this.v2 = this.New(ref...)
		})
	return this.v1, this.v2
}

func (this *Of2FnCached[T1, T2]) Value(ref ...any) (T1, T2) {
	this.mutex_condition_fn(func() bool { return this.done.Load() == 0 },
		func() {
			this.done.Store(1)
			this.v1, this.v2 = this.New(ref...)
		}, func() {
			if len(ref) > 0 {
				this.mutex_condition_fn(func() bool { return this.key != ref[0] },
					func() {

						this.done.Store(0)
					})
			}
		})
	return this.v1, this.v2
}

func (this *Of2Fn[T1, T2]) SetValue(v1 T1, v2 T2) {
	this.done.Store(1)
	this.v1, this.v2 = v1, v2
}
