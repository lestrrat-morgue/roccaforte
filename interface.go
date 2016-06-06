package roccaforte

import (
	"net/http"
	"time"

	"github.com/lestrrat/roccaforte/event"

	"google.golang.org/cloud/pubsub"
)

type Destination interface {
	Notify(event.Event) error
}

type Engine struct {
	Sources []EventSource
}

type EventSource interface {
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
	client *pubsub.Client
	outCh  chan event.Event
	Topic  string // PubSub topic name
}

type HTTPSource struct {
	http.Handler
	outCh  chan event.Event
	Listen string
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
