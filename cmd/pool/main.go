package pool

import "sync"

// Resetter объекты для сброса
type Resetter interface {
	Reset()
}

// Pool — пул объектов
type Pool[T any, PT Resetter] struct {
	internal sync.Pool
}

// New создает новый экземпляр пула
func New[T any, PT Resetter](factory func() PT) *Pool[T, PT] {
	return &Pool[T, PT]{
		internal: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

// Get извлекает объект из пула
func (p *Pool[T, PT]) Get() PT {
	return p.internal.Get().(PT)
}

// Put сбрасывает состояние объекта и возвращает его в пул
func (p *Pool[T, PT]) Put(obj PT) {
	obj.Reset()
	p.internal.Put(obj)
}
