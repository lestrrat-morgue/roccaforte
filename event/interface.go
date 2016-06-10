package event

type Event interface {
	ID() string
	Name() string
}

type CoreAttrs struct {
	id   string `json:"id"`
	name string `json:"name"`
}

type EventGroup struct {
	ID          int64
	Kind        string // value passed to datastore.NewQuery(kind)
	ProcessedOn int64  // non-zero if being processed by somebody
}
