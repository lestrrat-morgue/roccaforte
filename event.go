package roccaforte

import (
	"time"

	"github.com/builderscon/octav/octav/tools"
	"github.com/lestrrat/roccaforte/event"
)

func NewEvent(s EventSource, name string) event.Event {
	e := &ReceivedEvent{
		source:    s,
		createdOn: time.Now(),
	}
	e.CoreAttrs.SetName(name)
	e.CoreAttrs.SetID(tools.UUID())
	return e
}

func (e *ReceivedEvent) Source() EventSource {
	return e.source
}

func (e *ReceivedEvent) SetReceivedOn(t time.Time) {
	e.receivedOn = t
}

func (e *ReceivedEvent) SetDeliveredOn(t time.Time) {
	e.deliveredOn = t
}

func (e ReceivedEvent) ReceivedOn() time.Time {
	return e.receivedOn
}

func (e ReceivedEvent) DeliveredOn() time.Time {
	return e.deliveredOn
}