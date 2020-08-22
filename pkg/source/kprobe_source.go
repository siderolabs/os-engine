package source

import (
	bpf "github.com/iovisor/gobpf/bcc"

	"github.com/talos-systems/os-engine/pkg/events"
)

var _ Source = (*KprobeSource)(nil)

type KprobeSource struct {
	events.Streamer

	Module  *bpf.Module
	PerfCh  chan []byte
	PerfMap *bpf.PerfMap
}

func (s *KprobeSource) Load(stream events.Streamer) error {
	s.Streamer = stream

	return nil
}

func (s *KprobeSource) Start() error {
	s.PerfMap.Start()

	return nil
}

func (s *KprobeSource) Stop() error {
	s.PerfMap.Stop()
	s.Module.Close()

	return nil
}
