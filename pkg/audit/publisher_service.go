package audit

import "sync"

// PublisherService сервис подписчика
type PublisherService struct {
	mu        sync.RWMutex
	observers []Observer
	events    chan Event
}

// NewPublisherService создание нового сервиса
func NewPublisherService(bufferSize int) *PublisherService {
	return &PublisherService{
		observers: make([]Observer, 0),
		events:    make(chan Event, bufferSize),
	}
}

// Subscribe добавляет новый приемник в список
func (p *PublisherService) Subscribe(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, o)
}

// Notify теперь просто кладет событие в очередь
func (p *PublisherService) Notify(event Event) {
	p.events <- event
}

// Start запускает воркер, который слушает канал и уведомляет подписчиков
func (p *PublisherService) Start() {
	go func() {
		for event := range p.events {
			p.mu.RLock()
			for _, observer := range p.observers {
				observer.Publish(event)
			}
			p.mu.RUnlock()
		}
	}()
}
