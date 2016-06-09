package outgoing_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lestrrat/roccaforte/client"
	"github.com/lestrrat/roccaforte/event"
	"github.com/lestrrat/roccaforte/incoming"
	"github.com/lestrrat/roccaforte/outgoing"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var projectID string

func init() {
	projectID = os.Getenv("DATASTORE_PROJECT_ID")
}

func TestOutgoing(t *testing.T) {
	if projectID == "" {
		t.Skip("missing project ID. please set DATASTORE_PROJECT_ID")
		return
	}

	// Enqueue
	ctx, cancel := context.WithCancel(context.Background())

	h := incoming.NewHTTPSource()
	s := incoming.NewGDatastoreStorage(projectID)
	e := incoming.New()
	e.Storage = s
	e.AddSource(h)

	go func() {
		assert.NoError(t, e.Run(ctx), "incoming engine should exit w/o errors")
	}()

	go h.Loop(ctx)
	for i := 0; i < 10; i++ {
		e.SetRule("test.notify" + strconv.Itoa(i), &incoming.Rule{})
	}

	for i := 0; i < 10; i++ {
		go func() {
			cl := client.NewHTTP("http://localhost:8080/enqueue")
			for j := 0; j < 100; j++ {
				e := event.NewCoreAttrs("test.notify" + strconv.Itoa(i))
				if !assert.NoError(t, cl.Enqueue(e), "enqueue should succeed") {
					return
				}
			}
		}()
	}

	o := outgoing.New(projectID)
	go o.Run(ctx)

	time.AfterFunc(10*time.Second, cancel)
	<-ctx.Done()
}
