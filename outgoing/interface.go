package outgoing

import (
	"time"

	"github.com/WatchBeam/clock"
)

type Server struct {
	CheckInterval time.Duration
	Clock         clock.Clock
	ProjectID     string
}