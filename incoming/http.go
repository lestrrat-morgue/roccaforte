package incoming

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tylerb/graceful"
	"golang.org/x/net/context"
)

func NewHTTPSource() *HTTPSource {
	mux := http.NewServeMux()
	s := &HTTPSource{
		outCh:   make(chan []ReceivedEvent),
		Listen:  ":8080",
		Handler: mux,
	}
	mux.HandleFunc("/enqueue", s.httpEnqueue)

	return s
}

func (s *HTTPSource) SetStorage(es EventStorage) {
	s.storage = es
}

func (s *HTTPSource) Loop(ctx context.Context) {
	src := &graceful.Server{
		Server: &http.Server{
			Addr:    s.Listen,
			Handler: s.Handler,
		},
	}

	exited := make(chan struct{})

	go func() {
		defer close(exited)
		src.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		// Make sure to stop the server
		src.Stop(5 * time.Second)
	case <-exited:
	}

	// wait for the server to exit
	<-exited
}

func (s *HTTPSource) Events() <-chan []ReceivedEvent {
	return s.outCh
}

func httpError(w http.ResponseWriter, code int, msg string) {
	if msg == "" {
		msg = http.StatusText(code)
	}
	http.Error(w, msg, code)
}

func (s *HTTPSource) httpEnqueue(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if strings.ToLower(r.Method) != "post" {
		w.Header().Set("Allow", "POST")
		httpError(w, http.StatusMethodNotAllowed, "")
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		httpError(w, http.StatusBadRequest, "")
		return
	}

	events, err := s.toEvent(ctx, r)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.outCh <- events

	w.WriteHeader(http.StatusCreated)
}

func (s *HTTPSource) toEvent(ctx context.Context, r *http.Request) ([]ReceivedEvent, error) {
	now := time.Now()
	events := []ReceivedEvent{}
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		return nil, errors.Wrap(err, "failed to decode JSON")
	}

	// Save data (XXX This is too naive)
	for _, e := range events {
		e.SetReceivedOn(now)
		s.storage.Save(ctx, e)
	}

	return events, nil
}


