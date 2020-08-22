package loader

import (
	"fmt"
	"plugin"

	"github.com/talos-systems/os-engine/api/machinery"
	"github.com/talos-systems/os-engine/pkg/events"
	"github.com/talos-systems/os-engine/pkg/sink"
)

var _ sink.Sink = (*GoPluginSink)(nil)

type GoPluginSink struct {
	sink.Sink
}

func NewGoPluginSink(pluginpath string) (*GoPluginSink, error) {
	plug, err := plugin.Open(pluginpath)
	if err != nil {
		return nil, err
	}

	symb, err := plug.Lookup(machinery.SinkSymbol)
	if err != nil {
		return nil, err
	}

	if s, ok := symb.(sink.Sink); ok {
		return &GoPluginSink{Sink: s}, nil
	}

	return nil, fmt.Errorf("plugin does not implement sink interface")
}

func (d *GoPluginSink) Load(stream events.Streamer) error {
	return d.Sink.Load(stream)
}

func (g *GoPluginSink) Start() error {
	return g.Sink.Start()
}

func (g *GoPluginSink) Stop() error {
	return g.Sink.Stop()
}
