package audit

// Publisher интерфейс подписчика
type Publisher interface {
	Subscribe(o Observer)
	Notify(event Event)
	Start()
}
