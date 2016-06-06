package event

type Event interface {
	ID() string
	Name() string
}

type CoreAttrs struct {
	id   string `json:"id"`
	name string `json:"name"`
}
