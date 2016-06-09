package incoming_test

import (
	"os"
	"testing"
	"time"

	"github.com/lestrrat/roccaforte/incoming"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var projectID string

func init() {
	projectID = os.Getenv("DATASTORE_PROJECT_ID")
}

func TestGDatastore(t *testing.T) {
	if projectID == "" {
		t.Skip("missing project ID. please set DATASTORE_PROJECT_ID")
		return
	}
	ctx := context.Background()
	s := incoming.NewGDatastoreStorage(projectID)
	e := incoming.NewEvent(nil, "test.notify")
	e.SetReceivedOn(time.Now())
	if !assert.NoError(t, s.Save(ctx, 0, e), "s.Save should succeed") {
		return
	}
	if !assert.NoError(t, s.Delete(ctx, e), "s.Delete should succeed") {
		return
	}
}