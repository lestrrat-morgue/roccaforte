package incoming

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/lestrrat/roccaforte/internal/tools"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var eventType = reflect.TypeOf([]*ReceivedEvent{})

func New() *Server {
	return &Server{}
}

func (e *Server) SetRule(evname string, r *Rule) {
	e.Rules.Set(evname, r)
}

func (e *Server) AddSource(s EventSource) {
	s.SetStorage(e.Storage)
	e.Sources = append(e.Sources, s)
}

func (e *Server) Run(ctx context.Context) error {
	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	defer println("Terminating")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			println("Received signal")
			cancel()
			return
		}
	}()

	println("Serving requests...")
	cases := make([]reflect.SelectCase, len(e.Sources)+1)
	cases[0] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	for i, s := range e.Sources {
		cases[i+1] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(s.Events()),
		}
	}

	for loop := true; loop; {
		chosen, rv, ok := reflect.Select(cases)
		switch chosen {
		case 0:
			println("bail out")
			loop = false
			continue // ctx.Done
		default:
			if !ok {
				continue
			}

			if rv.Type() != eventType {
				return errors.New("received value was not a event.Event type (" + rv.Type().String() + ")")
			}

			if err := e.handleIncomingEvents(ctx, rv.Interface().([]*ReceivedEvent)); err != nil {
				if !tools.IsIgnorable(err) {
					return errors.Wrap(err, "failed to handle received value")
				}
			}
		}
	}

	return nil
}

func (s *Server) handleIncomingEvents(ctx context.Context, events []*ReceivedEvent) error {
	// Group by event names first, to make the processing faster
	byname := make(map[string][]*ReceivedEvent)
	for _, e := range events {
		byname[e.Name()] = append(byname[e.Name()], e)
	}

	var t int64
	for name, list := range byname {
		rule, err := s.Rules.Get(name)
		if err != nil {
			if !tools.IsIgnorable(err) {
				return errors.Wrap(err, "failed to lookup rule")
			}
			continue
		}

		if rule.Disabled() {
			continue
		}

		if t == 0 {
			// These events are all guaranteed to have the same received on date.
			t = list[0].ReceivedOn().Unix()
			if mod := t % rule.AggregationWindow(); mod > 0 {
				t = t - mod + rule.AggregationWindow()
			}
		}

		if err := s.Storage.Save(ctx, t, events...); err != nil {
			return errors.Wrap(err, "failed to add event for delivery")
		}
	}
	return nil
}
