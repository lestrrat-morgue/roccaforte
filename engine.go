package roccaforte

import (
	"reflect"

	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var eventType = reflect.TypeOf((*event.Event)(nil)).Elem()

func (e *Engine) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			continue // ctx.Done
		default:
			if !ok {
				continue
			}

			if rv.Type() != eventType {
				return errors.New("received value was not a event.Event type (" + rv.Type().String() + ")")
			}

			if err := e.handleEvent(rv.Interface().(event.Event)); err != nil {
				if !isIgnorable(err) {
					return errors.Wrap(err, "failed to handle received value")
				}
			}
		}
	}

	return nil
}

func (e *Engine) handleEvent(ev event.Event) error {
	//

	return nil
}
