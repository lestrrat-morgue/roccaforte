package incoming

import (
	"strconv"

	"golang.org/x/net/context"

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

func (s *GDatastoreStorage) Save(ctx context.Context, events ...*ReceivedEvent) error {
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
		key := datastore.NewIncompleteKey(ctx, e.Name(), parent)
		_, err := cl.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			var g int64
			if err := tx.Get(key, &g); err == nil {
				return nil
			}
			return tx.Put(key, id)
		}, nil)

		if err != nil {
			return errors.Wrap(err, "failed to store event id")
		}

		if _, err = cl.Put(ctx, key, e); err != nil {
			return errors.Wrap(err, "failed to Put event to datastore")
		}
	}
	return nil
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