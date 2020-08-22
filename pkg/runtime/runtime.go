package runtime

import (
	"github.com/talos-systems/os-engine/pkg/sink"
)

type Runtime interface {
	sink.Sink
}
