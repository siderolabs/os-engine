package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"unsafe"

	bpf "github.com/iovisor/gobpf/bcc"

	"github.com/talos-systems/os-engine/api/machinery"
	"github.com/talos-systems/os-engine/api/types"
	"github.com/talos-systems/os-engine/pkg/events"
	"github.com/talos-systems/os-engine/pkg/source"
)

import "C"

const bpfsource string = `
#include <linux/blkdev.h>

struct event_t {
	int major;
	int minor;
	char name[DISK_NAME_LEN];
	int op;
};

BPF_HASH(block, struct event_t);
BPF_PERF_OUTPUT(events);

int blk_add(struct pt_regs *ctx, struct device *parent, struct gendisk *disk,
			      const struct attribute_group **groups,
			      bool register_queue)
{
	struct event_t event = {};

	event.major = disk->major;
	event.minor = disk->first_minor;
	bpf_probe_read_kernel(event.name, sizeof(event.name), disk->disk_name);
	event.op = 1;

	events.perf_submit(ctx, &event, sizeof(event));

	bpf_trace_printk("added %s\n", event.name);

  return 0;
}

int blk_del(struct pt_regs *ctx, struct gendisk *disk)
{
	struct event_t event = {};

	event.major = disk->major;
	event.minor = disk->first_minor;
	bpf_probe_read_kernel(event.name, sizeof(event.name), disk->disk_name);
	event.op = 0;

	events.perf_submit(ctx, &event, sizeof(event));

	bpf_trace_printk("deleted %s\n", event.name);

  return 0;
}
`

type event struct {
	Major     int32
	Minor     int32
	Name      [32]byte
	Operation int32
}

var _ source.Source = (*blockdevice)(nil)

type blockdevice struct {
	*source.KprobeSource
}

var Source blockdevice

func (b *blockdevice) Load(stream events.Streamer) error {
	m := bpf.NewModule(bpfsource, []string{})

	add, err := m.LoadKprobe("blk_add")
	if err != nil {
		return fmt.Errorf("load blk_add: %w", err)
	}

	err = m.AttachKprobe("__device_add_disk", add, -1)
	if err != nil {
		return fmt.Errorf("attach __device_add_disk: %w", err)
	}

	del, err := m.LoadKprobe("blk_del")
	if err != nil {
		return fmt.Errorf("load blk_del: %w", err)
	}

	err = m.AttachKprobe("del_gendisk", del, -1)
	if err != nil {
		return fmt.Errorf("attach del_gendisk: %w", err)
	}

	table := bpf.NewTable(m.TableId("events"), m)

	perfCh := make(chan []byte)

	perfMap, err := bpf.InitPerfMap(table, perfCh, nil)
	if err != nil {
		return fmt.Errorf("init perf map: %w", err)
	}

	*b = blockdevice{
		KprobeSource: &source.KprobeSource{
			Streamer: stream,
			Module:   m,
			PerfCh:   perfCh,
			PerfMap:  perfMap,
		},
	}

	go Source.start()

	return nil
}

func (b *blockdevice) start() {
	log.Println("Blockdevice source started")

	var event event

	for {
		data := <-b.PerfCh

		err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
		if err != nil {
			b.Publish(&types.Error{
				Code:    1,
				Domain:  machinery.DomainStorage,
				Source:  machinery.SourceBlockdevice,
				Message: fmt.Sprintf("failed to decode data: %s", err),
			})

			continue
		}

		name := (*C.char)(unsafe.Pointer(&event.Name))

		var op string
		if event.Operation == 1 {
			op = "add"
		} else if event.Operation == 0 {
			op = "remove"
		}

		b.Publish(&types.Blockdevice{
			Name:      C.GoString(name),
			Major:     event.Major,
			Minor:     event.Minor,
			Operation: op,
		})
	}

	log.Println("Blockdevice source stopped")
}

// The main function can be used for local development and debugging.
func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	e := events.NewStream(1000, 10)
	Source.Load(e)

	Source.Start()

	defer Source.Stop()

	<-sig
}
