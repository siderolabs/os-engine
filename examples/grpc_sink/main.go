package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/talos-systems/os-engine/api/types"
	"google.golang.org/grpc"
)

type example struct{}

func (e *example) Events(stream types.Sink_EventsServer) error {
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}

		if err != nil {
			return err
		}

		fmt.Printf("gRPC Example: %+v\n", event)
	}
}

func main() {
	address := "/tmp/sink.sock"

	if _, err := os.Stat(address); err == nil {
		if err := os.Remove(address); err != nil {
			log.Fatal(err)
		}
	}

	lis, err := net.Listen("unix", address)
	if err != nil {
		log.Fatal(err)
	}

	e := &example{}

	s := grpc.NewServer()

	types.RegisterSinkServer(s, e)

	err = s.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
