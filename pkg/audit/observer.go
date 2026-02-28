package audit

// Observer интерфейс наблюдателя
type Observer interface {
	Publish(event Event) error
}
