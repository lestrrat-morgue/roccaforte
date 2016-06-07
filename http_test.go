package roccaforte_test

import (
	"testing"
	"time"

	"github.com/lestrrat/roccaforte"
	"github.com/lestrrat/roccaforte/client"
	"github.com/lestrrat/roccaforte/event"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestHTTPSource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// currently fails if port :8080 is not open
	s := roccaforte.NewHTTPSource()
	m := roccaforte.NewMemoryStorage()
	s.SetStorage(m)
	go s.Loop(ctx)
	time.AfterFunc(5*time.Second, func() {
		t.Logf("timeout reached")
		cancel()
	})

	msgcount := 1000
	go func() {
		cl := client.NewHTTP("http://localhost:8080/enqueue")

		for i := 0; i < msgcount/2; i++ {
			e := event.NewCoreAttrs("test.notify1")
			if !assert.NoError(t, cl.Enqueue(e), "enqueue should succeed") {
				return
			}
		}
	}()
	time.AfterFunc(2*time.Second, func() {
		cl := client.NewHTTP("http://localhost:8080/enqueue")

		events := make([]event.Event, msgcount/2)
		for i := 0; i < msgcount/2; i++ {
			events[i] = event.NewCoreAttrs("test.notify2")
		}

		if !assert.NoError(t, cl.Enqueue(events...), "enqueue should succeed") {
			return
		}
	})

	seen := make(map[string]struct{})

	count := 0
	for loop := true; loop; {
		select {
		case <-ctx.Done():
			loop = false
		case event := <-s.Events():
			t.Logf("new event: %s", event.ID())
			_, ok := seen[event.ID()]
			if !assert.False(t, ok, "Event must be new") {
				return
			}
			seen[event.ID()] = struct{}{}
			count++
			if count >= msgcount {
				loop = false
			}
		}
	}

	if !assert.Equal(t, msgcount, count, "msg count and processed count should be the same") {
		return
	}

	// All events should be in the same time frame
	var timeframe int64 = 0
	m.Walk(func(p int64, s  string, events []event.Event) {
		switch timeframe {
		case 0:
			timeframe = p
		default:
			if !assert.Equal(t, timeframe, p, "all event types should be in the same time frame") {
				return
			}
		}
		t.Logf("%d:%s %d events", p, s, len(events))
	})
}