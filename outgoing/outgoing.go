package outgoing

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WatchBeam/clock"
	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/cloud/datastore"
)

func New(id string) *Server {
	return &Server{
		CheckInterval: time.Minute,
		Clock:         clock.C,
		ProjectID:     id,
	}
}

func (s *Server) Run(ctx context.Context) error {
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

	cl, err := datastore.NewClient(ctx, s.ProjectID)
	if err != nil {
		return errors.Wrap(err, "failed create datastore client")
	}
	defer cl.Close()

	tick := s.Clock.NewTicker(s.CheckInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.Chan():
			// Look for event groups that are past the due date
			var groups []event.EventGroup
			var keys []*datastore.Key

			// Lookup groups and mark them as taken
			_, err := cl.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
				now := s.Clock.Now().Unix()

				var g event.EventGroup
				q := datastore.NewQuery("EventGroup").
					Filter("ID <=", now).
					Filter("ProcessedOn =", 0).
					Order("ID")

				for it := cl.Run(ctx, q); ; {
					key, err := it.Next(&g)
					if err != nil {
						break
					}

					g.ProcessedOn = now
					if _, err := tx.Put(key, &g); err != nil {
						return errors.Wrap(err, "failed to update event group")
					}
					groups = append(groups, g)
					keys = append(keys, key)
				}
				return nil
			})
			if err != nil {
				return errors.Wrap(err, "failed to fetch process group")
			}

			for i, g := range groups {
				// go process this ID
				go s.ProcessEventGroup(ctx, keys[i], g)
			}
		}
	}
	return nil
}

func (s *Server) ProcessEventGroup(ctx context.Context, key *datastore.Key, g event.EventGroup) error {
	defer s.RemoveEventGroup(ctx, key, g)

	// Now process the events
	/*
		rule, err := s.LookupRule(g.Kind)
		if err != nil {
			return errors.Wrap(err, "failed to lookup rule")
		}
	*/

	// Deliver. Note that we do not want to send notifications for
	// each event -- this is exzactly why we aggregated them in the first
	// place. Create a message, and send it.

	parent := datastore.NewKey(ctx, "ReceivedEvents", key.Name(), 0, nil)
	q := datastore.NewQuery(g.Kind).Ancestor(parent)

	cl, err := datastore.NewClient(ctx, s.ProjectID)
	if err != nil {
		return errors.Wrap(err, "failed create datastore client")
	}
	defer cl.Close()
	c, err := cl.Count(ctx, q)
	if err != nil {
		return errors.Wrap(err, "failed to get the count of events")
	}

	m := struct {
		ID         int64 `json:"id"`
		EventCount int   `json:"event_count"`
	}{
		ID:         key.ID(),
		EventCount: c,
	}

	json.NewEncoder(os.Stdout).Encode(m)

	return nil
}

func (s *Server) RemoveEventGroup(ctx context.Context, key *datastore.Key, g event.EventGroup) error {
	cl, err := datastore.NewClient(ctx, s.ProjectID)
	if err != nil {
		return errors.Wrap(err, "failed create datastore client")
	}
	defer cl.Close()

	parent := datastore.NewKey(ctx, "ReceivedEvents", key.Name(), 0, nil)
	q := datastore.NewQuery(g.Kind).
		Ancestor(parent).
		KeysOnly().
		Limit(1000)
	keys := make([]*datastore.Key, 1000)
	for {
		keys = keys[0:0]
		for it := cl.Run(ctx, q); ; {
			key, err := it.Next(nil)
			if err != nil {
				break
			}
			keys = append(keys, key)
		}
		if len(keys) == 0 {
			break
		}
		cl.DeleteMulti(ctx, keys)
	}

	cl.Delete(ctx, key)
	return nil
}
