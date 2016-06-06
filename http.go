package roccaforte

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"
	"github.com/tylerb/graceful"
	"golang.org/x/net/context"
)

func NewHTTPSource() *HTTPSource {
	mux := http.NewServeMux()
	s := &HTTPSource{
		outCh:   make(chan event.Event),
		Listen:  ":8080",
		Handler: mux,
	}
	mux.HandleFunc("/enqueue", s.httpEnqueue)

	return s
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

func (s *HTTPSource) Events() <-chan event.Event {
	return s.outCh
}

func httpError(w http.ResponseWriter, code int, msg string) {
	if msg == "" {
		msg = http.StatusText(code)
	}
	http.Error(w, msg, code)
}

func (s *HTTPSource) httpEnqueue(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Method) != "post" {
		w.Header().Set("Allow", "POST")
		httpError(w, http.StatusMethodNotAllowed, "")
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		httpError(w, http.StatusBadRequest, "")
		return
	}

	events, err := s.toEvent(r)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, e := range events {
		// Blocking here... is it a good idea?
		s.outCh <- e
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *HTTPSource) toEvent(r *http.Request) ([]ReceivedEvent, error) {
	events := []ReceivedEvent{}
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		return nil, errors.Wrap(err, "failed to decode JSON")
	}
	return events, nil
}
