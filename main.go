package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/talos-systems/os-engine/internal/controller"
	"github.com/talos-systems/os-engine/internal/loader"
	"github.com/talos-systems/os-engine/pkg/events"
	"github.com/talos-systems/os-engine/pkg/sink"
)

type Options struct {
	Runtime string
	Plugins string
	Socket  string
	Debug   bool
}

var options Options

func init() {
	flag.StringVar(&options.Runtime, "runtime", "/lib/engine/runtime/runtime.so", "the path to the runtime")
	flag.StringVar(&options.Plugins, "plugins-dir", "/lib/engine/plugins", "the directory to load plugins from")
	flag.StringVar(&options.Socket, "sink-socket", "", "the path to a local gRPC unix socket")
	flag.BoolVar(&options.Debug, "debug", false, "enable the debug sink")
}

func main() {
	flag.Parse()

	plugins, err := loader.LoadPlugins(options.Runtime, options.Plugins)
	if err != nil {
		log.Fatal(err)
	}

	if options.Debug {
		plugins.Sinks = append(plugins.Sinks, &sink.DebugSink{})
	}

	if options.Socket != "" {
		sink, err := loader.NewGRPCSink(options.Socket)
		if err != nil {
			log.Fatal(err)
		}

		plugins.Sinks = append(plugins.Sinks, sink)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	stream := events.NewStream(1000, 10)

	con := controller.NewController(stream)

	err = con.Start(plugins)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("All sources and sinks started")

	<-sig

	con.Stop()
}
