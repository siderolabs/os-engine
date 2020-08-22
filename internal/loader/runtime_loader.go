package loader

import (
	"fmt"
	"plugin"

	"github.com/talos-systems/os-engine/api/machinery"
	"github.com/talos-systems/os-engine/pkg/runtime"
)

type Runtime struct {
	runtime.Runtime
}

func LoadRuntime(pluginpath string) (*Runtime, error) {
	plug, err := plugin.Open(pluginpath)
	if err != nil {
		return nil, err
	}

	symb, err := plug.Lookup(machinery.RuntimeSymbol)
	if err != nil {
		return nil, err
	}

	if r, ok := symb.(runtime.Runtime); ok {
		return &Runtime{r}, nil
	}

	return nil, fmt.Errorf("plugin does not implement runtime interface")
}
