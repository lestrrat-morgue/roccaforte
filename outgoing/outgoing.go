package outgoing

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/cloud/datastore"
)

func New(id string) *Server {
	return &Server{
		ProjectID: id,
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

	tick := time.NewTicker(time.Minute)
	defer tick.Stop()

	// Look for event groups that are past the due date
	select {
	case <-ctx.Done():
		return nil
	case <-tick.C:
		var g event.EventGroup
		q := datastore.NewQuery("EventGroup").
			Filter("ID <=", time.Now().Unix()).
			Order("ID")

		for it := cl.Run(ctx, q); ; {
			key, err := it.Next(&g)
			if err != nil {
				break
			}

			// go process this ID
			go s.ProcessEventGroup(ctx, key, g)
		}
	}
	return nil
}

func (s *Server) ProcessEventGroup(ctx context.Context, key *datastore.Key, g event.EventGroup) error {
	defer s.RemoveEventGroup(ctx, key, g)

	// do stuff
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
