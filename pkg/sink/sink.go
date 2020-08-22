package sink

import (
	"github.com/talos-systems/os-engine/internal/unit"
	"github.com/talos-systems/os-engine/pkg/events"
)

type Sink interface {
	unit.Unit
	events.Watcher
}
