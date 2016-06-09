package incoming

import (
	"strconv"

	"golang.org/x/net/context"

	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"
	"google.golang.org/cloud/datastore"
)

func NewGDatastoreStorage(projectID string) *GDatastoreStorage {
	return &GDatastoreStorage{
		ProjectID: projectID,
	}
}

func (s *GDatastoreStorage) client(ctx context.Context) (*datastore.Client, error) {
	return datastore.NewClient(ctx, s.ProjectID)
}

func (s *GDatastoreStorage) Save(ctx context.Context, fireT int64, events ...*ReceivedEvent) error {
	cl, err := s.client(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create datastore client")
	}

	// get distinct event names
	eventNames := make(map[string]struct{})
	for _, e := range events {
		if _, ok := eventNames[e.Name()]; ok {
			continue
		}
		eventNames[e.Name()] = struct{}{}
	}


	parent := datastore.NewKey(ctx, "ReceivedEvents", strconv.FormatInt(fireT, 10), 0, nil)
	keys := make([]*datastore.Key, len(events))
	// classify entries into basetime / event name
	for i, e := range events {
		keys[i] = datastore.NewIncompleteKey(ctx, e.Name(), parent)
	}

	if _, err = cl.PutMulti(ctx, keys, events); err != nil {
		return errors.Wrap(err, "failed to execute PutMulti")
	}

	// Create EventGroup for fire time `fireT`
	// Create these at the end, because without these, there will be no events fired
	_, err = cl.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		for name := range eventNames {
			k := datastore.NewKey(ctx, "EventGroup", strconv.FormatInt(fireT, 10), 0, nil)
			var g event.EventGroup
			if err := tx.Get(k, &g); err == nil {
				continue
			}
			g.ID = fireT
			g.Kind = name
			tx.Put(k, &g)
		}
		return nil
	})

	return errors.Wrap(err, "failed to create event group")
}

func (s *GDatastoreStorage) Delete(ctx context.Context, events ...*ReceivedEvent) error {
	cl, err := s.client(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create datastore client")
	}

	// classify entries into basetime / event name
	for _, e := range events {
		id := e.ReceivedOn().Unix()
		if mod := id % 300; mod > 0 {
			id = id - mod
		}

		parent := datastore.NewKey(ctx, "ReceivedEvents", strconv.FormatInt(id, 10), 0, nil)
		q := datastore.NewQuery(e.Name()).
			Ancestor(parent).
			Filter("ID = ", e.ID()).
			KeysOnly()
		for it := cl.Run(ctx, q); ; {
			key, err := it.Next(nil)
			if err != nil {
				break
			}
			if err := cl.Delete(ctx, key); err != nil {
				return errors.Wrap(err, "failed to Delete event from datastore")
			}
		}
	}
	return nil
}