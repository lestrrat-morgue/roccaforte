package incoming_test

import (
	"testing"

	"github.com/lestrrat/roccaforte/incoming"
	"golang.org/x/net/context"
)

func TestGDatastore(t *testing.T) {
	s := incoming.NewGDatastoreStorage("")
	s.Save(context.Background(), incoming.NewEvent(nil, "test.notify"))
}