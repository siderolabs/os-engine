package loader

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/talos-systems/os-engine/pkg/sink"
	"github.com/talos-systems/os-engine/pkg/source"
)

type Plugins struct {
	Sources []source.Source
	Sinks   []sink.Sink
}

func LoadPlugins(runtimepath, dir string) (plugins *Plugins, err error) {
	plugins = &Plugins{}

	if runtimepath != "" {
		runtime, err := LoadRuntime(runtimepath)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Loaded %q", runtimepath)

		plugins.Sinks = append(plugins.Sinks, runtime)
	}

	if dir == "" {
		return plugins, nil
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		path := filepath.Join(dir, file.Name())

		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}

		source, err := NewGoPluginSource(abs)
		if err == nil {
			log.Printf("Loaded %q", abs)

			plugins.Sources = append(plugins.Sources, source)

			continue
		}

		sink, err := NewGoPluginSink(abs)
		if err == nil {
			log.Printf("Loaded %q", abs)

			plugins.Sinks = append(plugins.Sinks, sink)

			continue
		}

		log.Printf("Failed to load plugin: %s", path)
	}

	return plugins, nil
}
