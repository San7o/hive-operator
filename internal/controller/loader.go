package controller

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	RingbuffReader *ringbuf.Reader = nil
	Objs           bpfObjects      = bpfObjects{}
	KeyProbe       link.Link       = nil
)

func LoadBpf(ctx context.Context) error {
	kprobed_func := "inode_permission"

	log := log.FromContext(ctx)
	// Remove resource limits for kernels <5.11.
	err := rlimit.RemoveMemlock()
	if err != nil {
		log.Error(err, "Error Removing memlock")
		return err
	}

	if err = loadBpfObjects(&Objs, nil); err != nil {
		log.Error(err, "Error nLoading eBPF objects")
		return err
	}

	KeyProbe, err = link.Kprobe(kprobed_func, Objs.KprobeInodePermission, nil)
	if err != nil {
		log.Error(err, "Error Opening kprobe")
		return err
	}

	// Fill the map with some data
	var ino uint64 = 123
	var ino2 uint64 = 1234
	var key0 uint32 = 0
	var key1 uint32 = 1
	err = Objs.TracedInodes.Update(key0, ino, ebpf.UpdateAny)
	if err != nil {
		log.Error(err, "Error Updating map")
		return err
	}
	err = Objs.TracedInodes.Update(key1, ino2, ebpf.UpdateAny)
	if err != nil {
		log.Error(err, "Error Updating map")
		return err
	}

	// Open a ringbuf reader from userspace RINGBUF map described in the
	// eBPF C program.
	RingbuffReader, err = ringbuf.NewReader(Objs.Rb)
	if err != nil {
		log.Error(err, "Error opening ringbuf reader")
		return err

	}
	return nil
}

func LogData(ctx context.Context) {

	log := log.FromContext(ctx)

	if RingbuffReader == nil {
		log.Error(nil, "Logger error: ringbuffer not inizialized")
		return
	}

	var data bpfLogData
	for {
		record, err := RingbuffReader.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				log.Error(err, "Exiting logger: buffer closed.")
				return
			}
			continue
		}

		// Parse the ringbuf data entry into a logData structure.
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &data); err != nil {
			log.Error(err, "Error parsing ringbuf data")
			continue
		}

		log.Info("New event", "pid", data.Pid, "tgid", data.Tgid, "uid",
			data.Uid, "gid", data.Gid, "ino", data.Ino, "mask", data.Mask)
	}
}
