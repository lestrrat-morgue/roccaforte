package roccaforte

import (
	"encoding/json"
	"time"

	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"

	"golang.org/x/net/context"
	"google.golang.org/cloud/pubsub"
)

func NewGPubSubSource(cl *pubsub.Client) *GPubSubSource {
	return &GPubSubSource{
		client: cl,
		outCh:  make(chan event.Event),
	}
}

func (s *GPubSubSource) Events() <-chan event.Event {
	return s.outCh
}

func (s *GPubSubSource) Loop(ctx context.Context) {
	backoff := time.Second
	sub := s.client.Subscription(s.Topic)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		iter, err := sub.Pull(ctx)
		if err != nil {
			// oh, error, eh.
			time.Sleep(backoff)
			if backoff < 3*time.Minute {
				backoff = time.Duration(int(float64(backoff) * 1.2))
			}
			continue
		}

		// reset exponential backoff
		backoff = time.Second

		for {
			payload, err := iter.Next()
			if err != nil {
				break
			}

			s.handlePayload(payload)
			payload.Done(true)
		}
	}
}

func (s *GPubSubSource) handlePayload(payload *pubsub.Message) error {
	events := []ReceivedEvent{}
	if err := json.Unmarshal(payload.Data, &events); err != nil {
		return errors.Wrap(err, "failed to deserialize JSON event data")
	}

	for _, e := range events {
		s.outCh <- e
	}
	return nil
}
