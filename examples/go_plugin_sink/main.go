package main

import (
	"fmt"

	"github.com/talos-systems/os-engine/pkg/events"
	"github.com/talos-systems/os-engine/pkg/sink"
)

var _ sink.Sink = (*example)(nil)

type example struct {
	events.Streamer
}

var Sink example

func (e *example) Load(stream events.Streamer) error {
	e.Streamer = stream

	return nil
}

func (e *example) Start() error {
	return e.Watch(func(events <-chan events.Event) {
		for {
			event := <-events

			fmt.Printf("Go Plugin Example: %+v\n", event)
		}
	})
}

func (*example) Stop() error {
	return nil
}
