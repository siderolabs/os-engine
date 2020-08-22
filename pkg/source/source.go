package source

import (
	"github.com/talos-systems/os-engine/internal/unit"
	"github.com/talos-systems/os-engine/pkg/events"
)

type Source interface {
	unit.Unit
	events.Publisher
}
