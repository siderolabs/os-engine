package sink

import (
	"fmt"

	"github.com/talos-systems/os-engine/pkg/events"
)

var _ Sink = (*DebugSink)(nil)

type DebugSink struct {
	events.Streamer
}

func (d *DebugSink) Load(stream events.Streamer) error {
	d.Streamer = stream

	return nil
}

func (d *DebugSink) Start() error {
	return d.Watch(func(events <-chan events.Event) {
		for {
			event := <-events

			fmt.Printf("Debug: %+v\n", event)
		}
	})
}

func (*DebugSink) Stop() error {
	return nil
}
