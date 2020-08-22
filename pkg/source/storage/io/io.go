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
#include <uapi/linux/ptrace.h>
#include <linux/blkdev.h>

struct val_t {
    u64 ts;
    u32 pid;
    char name[TASK_COMM_LEN];
};

struct event_t {
    u32 pid;
    int rwflag;
    u64 delta;
    u64 qdelta;
    u64 sector;
    u64 len;
    u64 ts;
    char disk_name[DISK_NAME_LEN];
    char name[TASK_COMM_LEN];
};

BPF_HASH(start, struct request *);
BPF_HASH(infobyreq, struct request *, struct val_t);
BPF_PERF_OUTPUT(events);

int trace_pid_start(struct pt_regs *ctx, struct request *req)
{
    struct val_t val = {};
		u64 ts;

    if (bpf_get_current_comm(&val.name, sizeof(val.name)) == 0) {
        val.pid = bpf_get_current_pid_tgid() >> 32;
        if (1) {
            val.ts = bpf_ktime_get_ns();
        }
        infobyreq.update(&req, &val);
		}

    return 0;
}

int trace_req_start(struct pt_regs *ctx, struct request *req)
{
		u64 ts;

    ts = bpf_ktime_get_ns();
		start.update(&req, &ts);

    return 0;
}

int trace_req_completion(struct pt_regs *ctx, struct request *req)
{
    u64 *tsp;
    struct val_t *valp;
    struct event_t data = {};
		u64 ts;

    // fetch timestamp and calculate delta
		tsp = start.lookup(&req);

    if (tsp == 0) {
        // missed tracing issue
        return 0;
		}

    ts = bpf_ktime_get_ns();
    data.delta = ts - *tsp;
    data.ts = ts / 1000;
    data.qdelta = 0;
		valp = infobyreq.lookup(&req);

    if (valp == 0) {
        data.len = req->__data_len;
        data.name[0] = '?';
        data.name[1] = 0;
    } else {
        if (1) {
            data.qdelta = *tsp - valp->ts;
				}

        data.pid = valp->pid;
        data.len = req->__data_len;
        data.sector = req->__sector;
        bpf_probe_read_kernel(&data.name, sizeof(data.name), valp->name);
        struct gendisk *rq_disk = req->rq_disk;
        bpf_probe_read_kernel(&data.disk_name, sizeof(data.disk_name),
                       rq_disk->disk_name);
		}

    data.rwflag = !!((req->cmd_flags & REQ_OP_MASK) == REQ_OP_WRITE);
    events.perf_submit(ctx, &data, sizeof(data));
    start.delete(&req);
    infobyreq.delete(&req);
    return 0;
}
`

type event struct {
	Pid      uint32
	Rwflag   int32
	Delta    uint64
	Qdelta   uint64
	Sector   uint64
	Len      uint64
	Ts       uint64
	DiskName [32]byte
	Name     [16]byte
}

var _ source.Source = (*io)(nil)

type io struct {
	*source.KprobeSource
}

var Source io

func (i *io) Load(stream events.Streamer) error {
	m := bpf.NewModule(bpfsource, []string{})

	tracePidStart, err := m.LoadKprobe("trace_pid_start")
	if err != nil {
		return fmt.Errorf("load trace_pid_start: %w", err)
	}

	err = m.AttachKprobe("blk_account_io_start", tracePidStart, -1)
	if err != nil {
		return fmt.Errorf("attach blk_account_io_start: %w", err)
	}

	traceReqStart, err := m.LoadKprobe("trace_req_start")
	if err != nil {
		return fmt.Errorf("load trace_req_start: %w", err)
	}

	err = m.AttachKprobe("blk_mq_start_request", traceReqStart, -1)
	if err != nil {
		return fmt.Errorf("attch blk_mq_start_request: %w", err)
	}

	traceReqCompletion, err := m.LoadKprobe("trace_req_completion")
	if err != nil {
		return fmt.Errorf("load trace_req_completion: %w", err)
	}

	err = m.AttachKprobe("blk_account_io_done", traceReqCompletion, -1)
	if err != nil {
		return fmt.Errorf("attach blk_account_io_done: %w", err)

	}

	table := bpf.NewTable(m.TableId("events"), m)

	perfCh := make(chan []byte)

	perfMap, err := bpf.InitPerfMap(table, perfCh, nil)
	if err != nil {
		return fmt.Errorf("init perf map: %w", err)
	}

	*i = io{
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

func (i *io) start() {
	log.Println("IO source started")

	var event event

	for {
		data := <-i.PerfCh

		err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
		if err != nil {
			i.Publish(&types.Error{
				Code:    1,
				Domain:  machinery.DomainStorage,
				Source:  machinery.SourceIO,
				Message: fmt.Sprintf("failed to decode data: %s", err),
			})

			continue
		}

		cmd := (*C.char)(unsafe.Pointer(&event.Name))
		name := (*C.char)(unsafe.Pointer(&event.DiskName))

		var op string
		if event.Rwflag == 1 {
			op = "write"
		} else if event.Rwflag == 0 {
			op = "read"
		}

		i.Publish(&types.IO{
			Command:   C.GoString(cmd),
			Pid:       event.Pid,
			Disk:      C.GoString(name),
			Operation: op,
		})
	}

	log.Println("IO source stopped")
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
