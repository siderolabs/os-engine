package unit

import "github.com/talos-systems/os-engine/pkg/events"

type Unit interface {
	Load(events.Streamer) error
	Start() error
	Stop() error
}
