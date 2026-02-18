package audit

// NoopPublisher мок издателя
type NoopPublisher struct{}

// NewNoopPublisher создание мока
func NewNoopPublisher() Publisher {
	return &NoopPublisher{}
}

// Subscribe подписывает
func (p *NoopPublisher) Subscribe(o Observer) {
}

// Notify уведомляет
func (p *NoopPublisher) Notify(e Event) {
}

// Start начало процесса
func (p *NoopPublisher) Start() {
}
