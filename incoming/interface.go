package incoming

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/lestrrat/roccaforte/event"

	"google.golang.org/cloud/pubsub"
)

type Rule struct {}

type Destination interface {
	Notify(event.Event) error
}

type Server struct {
	Sources []EventSource
	Storage EventStorage
}

type EventStorage interface {
	Save(context.Context, ...event.Event) error
	AddEventDelivery(int64, []ReceivedEvent) error
}

type EventSource interface {
	SetStorage(EventStorage)
	Events() <-chan event.Event
}

type ReceivedEvent struct {
	event.CoreAttrs

	source EventSource
	// time this event was first created
	createdOn time.Time
	// time this event was received by the processor
	receivedOn time.Time
	// time this event was finally completely delivered
	deliveredOn time.Time
}

type GPubSubSource struct {
	client  *pubsub.Client
	outCh   chan []ReceivedEvent
	storage EventStorage
	Topic   string // PubSub topic name
}

type HTTPSource struct {
	http.Handler
	outCh   chan []ReceivedEvent
	storage EventStorage
	Listen  string
}

/*
// BundledEvent bundles multiple events that arrived in a particular
// time frame
type BundledEvent struct {

}

// Destination is where notifications get delivered to
type Destination interface {
	Deliver(Notification)
}*/

type MemoryStorage struct {
	mutex sync.Mutex
	store map[int64]map[string][]event.Event
}

type GDatastoreStorage struct {
	ProjectID string
}
