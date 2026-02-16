package audit

type Publisher interface {
	Subscribe(o Observer)
	Notify(event Event)
	Start()
}
