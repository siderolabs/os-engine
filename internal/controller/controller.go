package controller

import (
	"log"
	"sync"

	"github.com/talos-systems/os-engine/internal/loader"
	"github.com/talos-systems/os-engine/pkg/events"
)

type Controller interface {
	Start(*loader.Plugins) error
	Stop() error
}

var _ Controller = (*Default)(nil)

type Default struct {
	stream  events.Streamer
	plugins *loader.Plugins
}

func NewController(stream events.Streamer) *Default {
	return &Default{stream: stream}
}

func (d *Default) Start(plugins *loader.Plugins) error {
	d.plugins = plugins

	var wg sync.WaitGroup

	wg.Add(len(d.plugins.Sources))

	for _, source := range d.plugins.Sources {
		source := source

		go func() {
			defer wg.Done()

			err := source.Load(d.stream)
			if err != nil {
				log.Println(err)
			}

			err = source.Start()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	wg.Add(len(d.plugins.Sinks))

	for _, sink := range d.plugins.Sinks {
		sink := sink

		go func() {
			defer wg.Done()

			err := sink.Load(d.stream)
			if err != nil {
				log.Println(err)
			}

			err = sink.Start()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	wg.Wait()

	return nil
}

func (d *Default) Stop() error {
	var wg sync.WaitGroup

	wg.Add(len(d.plugins.Sources))

	for _, source := range d.plugins.Sources {
		source := source

		go func() {
			defer wg.Done()

			err := source.Stop()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	wg.Add(len(d.plugins.Sinks))

	for _, sink := range d.plugins.Sinks {
		sink := sink

		go func() {
			defer wg.Done()

			err := sink.Stop()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	wg.Wait()

	return nil
}
