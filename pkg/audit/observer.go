package audit

type Observer interface {
	Publish(event Event) error
}
