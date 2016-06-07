package roccaforte

import (
	"github.com/lestrrat/roccaforte/event"
	"golang.org/x/net/context"
)

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[int64]map[string][]event.Event),
	}
}

func (ms *MemoryStorage) Save(ctx context.Context, events ...event.Event) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, e := range events {
		key := e.(ReceivedEvent).ReceivedOn().Unix()
		if mod := key % 300; mod > 0 {
			key = key - mod
		}
		byEvent, ok := ms.store[key]
		if !ok {
			byEvent = make(map[string][]event.Event)
			ms.store[key] = byEvent
		}
		byEvent[e.Name()] = append(byEvent[e.Name()], e)
	}
	return nil
}

func (ms *MemoryStorage) Walk(f func(int64, string, []event.Event)) {
	for t, em := range ms.store {
		for name, events := range em {
			f(t, name, events)
		}
	}
}
