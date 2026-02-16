package audit

type NoopPublisher struct{}

func NewNoopPublisher() Publisher {
	return &NoopPublisher{}
}

func (p *NoopPublisher) Subscribe(o Observer) {
}

func (p *NoopPublisher) Notify(e Event) {
}

func (p *NoopPublisher) Start() {
}
