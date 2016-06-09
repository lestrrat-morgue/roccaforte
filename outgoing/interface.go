package outgoing

import (
	"github.com/WatchBeam/clock"
)

type Server struct {
	Clock     clock.Clock
	ProjectID string
}