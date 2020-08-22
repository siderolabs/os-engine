package loader

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/talos-systems/os-engine/api/types"
	"github.com/talos-systems/os-engine/pkg/events"
	"github.com/talos-systems/os-engine/pkg/sink"
	"google.golang.org/grpc"
)

var _ sink.Sink = (*GRPCSink)(nil)

type GRPCSink struct {
	events.Streamer

	conn   *grpc.ClientConn
	stream types.Sink_EventsClient
}

func NewGRPCSink(address string) (*GRPCSink, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		return nil, err
	}

	client := types.NewSinkClient(conn)

	stream, err := client.Events(context.Background())
	if err != nil {
		return nil, err
	}

	g := &GRPCSink{
		conn:   conn,
		stream: stream,
	}

	return g, nil
}

func (g *GRPCSink) Load(stream events.Streamer) error {
	g.Streamer = stream

	return nil
}

func (g *GRPCSink) Start() error {
	return g.Watch(func(events <-chan events.Event) {
		for {
			event := <-events

			t, err := event.ToAPIEvent()
			if err != nil {
				log.Println(err)
			}

			err = g.stream.Send(t)
			if err != nil {
				log.Println(err)
			}
		}
	})

}

func (g *GRPCSink) Stop() error {
	return g.conn.Close()
}
