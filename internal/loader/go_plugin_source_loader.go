package loader

import (
	"fmt"
	"plugin"

	"github.com/talos-systems/os-engine/api/machinery"
	"github.com/talos-systems/os-engine/pkg/source"
)

var _ source.Source = (*GoPluginSource)(nil)

type GoPluginSource struct {
	source.Source
}

func NewGoPluginSource(pluginpath string) (*GoPluginSource, error) {
	plug, err := plugin.Open(pluginpath)
	if err != nil {
		return nil, err
	}

	symb, err := plug.Lookup(machinery.SourceSymbol)
	if err != nil {
		return nil, err
	}

	if s, ok := symb.(source.Source); ok {
		return &GoPluginSource{Source: s}, nil
	}

	return nil, fmt.Errorf("plugin does not implement source interface")
}
