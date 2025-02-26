package lazy

import (
	"sync"
	"sync/atomic"
)

type Of[T any] struct {
	done  atomic.Uint32
	m     sync.Mutex
	New   func() T
	value T
}

func (this *Of[T]) Value() T {
	if this.done.Load() == 0 {
		func() {
			this.m.Lock()
			defer this.m.Unlock()
			if this.done.Load() == 0 {
				this.done.Store(1)
				this.value = this.New()
			}
		}()
	}
	return this.value
}

func Make[T any](newfunc func() T) Of[T] {
	return Of[T]{New: newfunc}
}

func New[T any](newfunc func() T) *Of[T] {
	return &Of[T]{New: newfunc}
}
