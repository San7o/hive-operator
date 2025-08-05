/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package ebpf

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kivev1alpha1 "github.com/San7o/kivebpf/api/v1alpha1"
	container "github.com/San7o/kivebpf/internal/controller/container"
)

const (
	MapMaxEntries = 1024
	KprobedFunc   = "inode_permission"
)

var (
	RingbuffReader *ringbuf.Reader = nil
	Objs           bpfObjects      = bpfObjects{}
	Kprobe         link.Link       = nil
	Loaded         bool            = false
)

/*
 *  Loads the eBPF objects, the eBPF program and opens the ring
 *  buffer.
 */
func LoadEbpf(ctx context.Context) error {

	// Remove resource limits for kernels <5.11.
	err := rlimit.RemoveMemlock()
	if err != nil {
		return fmt.Errorf("LoadEbpf Error Remove memlock: %w", err)
	}

	if err = loadBpfObjects(&Objs, nil); err != nil {
		return fmt.Errorf("LoadEbpf Error Load eBPF objects: %w", err)
	}

	Kprobe, err = link.Kprobe(KprobedFunc, Objs.KprobeInodePermission, nil)
	if err != nil {
		return fmt.Errorf("LoadEbpf Error Open kprobe: %w", err)
	}

	RingbuffReader, err = ringbuf.NewReader(Objs.Rb)
	if err != nil {
		return fmt.Errorf("LoadEbpf Error Open ringbuf reader: %w", err)
	}

	Loaded = true
	return nil
}

/*
 *  Unload the eBPF program, objects and ringbuffer.
 */
func UnloadEbpf(ctx context.Context) error {

	if Kprobe != nil {

		if err := Kprobe.Close(); err != nil {
			return fmt.Errorf("UnloadEbpf Error Failed to close ebpf program: %w", err)
		}

		if err := Objs.TracedInodes.Close(); err != nil {
			return fmt.Errorf("UnloadEbpf Error Failed to close eBPF map: %w", err)
		}

		if err := Objs.Close(); err != nil {
			return fmt.Errorf("UnloadEbpf Error Failed to close eBPF objects: %w", err)
		}

		if RingbuffReader != nil {
			if err := RingbuffReader.Close(); err != nil {
				return fmt.Errorf("UnloadEbpf Error failed to close the Rinbuffer Reader: %w", err)
			}
		}

		Loaded = false
	}

	return nil
}

/*
 *  Read the data from the Ringbuffer, hangs until data is received or
 *  returns an error. This function can be used without a running
 *  kubernetes cluster.
 */
func ReadEbpfData() (bpfLogData, error) {

	var data bpfLogData
	record, err := RingbuffReader.Read()
	if err != nil {
		if errors.Is(err, ringbuf.ErrClosed) {
			return bpfLogData{}, fmt.Errorf("ReadAlert Error Buffer closed.")
		}
	}

	err = binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &data)
	if err != nil {
		return bpfLogData{}, fmt.Errorf("ReadAlert Error Parse ringbuf data")
	}

	return data, nil
}

func ReadAlert(ctx context.Context, cli client.Reader) (kivev1alpha1.KiveAlert, error) {

	if RingbuffReader == nil {
		return kivev1alpha1.KiveAlert{}, fmt.Errorf("ReadAlert Error Ringbuffer not inizialized")
	}

	log := log.FromContext(ctx)
	data, err := ReadEbpfData() // Hangs
	if err != nil {
		return kivev1alpha1.KiveAlert{}, fmt.Errorf("ReadAlert Error Reading Ebpf Data: %w", err)
	}

	kiveDataList := &kivev1alpha1.KiveDataList{}
	err = cli.List(ctx, kiveDataList)
	if err != nil {
		return kivev1alpha1.KiveAlert{}, fmt.Errorf("ReadAlert Error Failed to get Kive Data resource: %w", err)
	}

	for _, kiveData := range kiveDataList.Items {
		if kiveData.Spec.InodeNo == data.Ino {

			cwd := ""
			// If the node is a container on some host, then we need to read
			// the host's procfs which is assumed to be mounted in
			// /host/real/proc. If this does not exist, either the cluster
			// is misconfigured or it is non containerized, then we check
			// the regualr procfs of the node. This workaround is needed
			// since procfs of a node is not the same as the host if the
			// node is a container on the host (for example, for clusters
			// created using Kind)
			cwd, err = os.Readlink(fmt.Sprintf("%s/%d/cwd", container.RealHostProcMountpoint, data.Pid))
			if err != nil {
				cwd, _ = os.Readlink(fmt.Sprintf("%s/%d/cwd", container.ProcMountpoint, data.Pid))
				// error is handled gracefully
				log.Info(fmt.Sprintf("Could not read %s/%d/cwd while generating an KiveAlert, this can happen if the process terminated too quickly for the operator to react or the node is running in a container and procfs is not mounted in %s", container.ProcMountpoint, data.Pid, container.RealHostProcMountpoint))
			}
			out := kivev1alpha1.KiveAlert{
				Timestamp:      time.Now().Format(time.RFC3339),
				KivePolicyName: kiveData.Annotations["kive_policy_name"],
				Metadata: kivev1alpha1.KiveAlertMetadata{
					Path:     kiveData.Annotations["path"],
					Inode:    data.Ino,
					Mask:     data.Mask,
					KernelID: kiveData.Spec.KernelID,
					Callback: kiveData.ObjectMeta.Annotations["callback"],
				},
				Pod: kivev1alpha1.PodMetadata{
					Name:      kiveData.Annotations["pod_name"],
					Namespace: kiveData.Annotations["namespace"],
					Container: kivev1alpha1.ContainerMetadata{
						Id:   kiveData.Annotations["container_id"],
						Name: kiveData.Annotations["container_name"],
					},
					Ip: kiveData.Annotations["ip"],
				},
				Node: kivev1alpha1.NodeMetadata{
					Name: kiveData.Annotations["node_name"],
				},
				Process: kivev1alpha1.ProcessMetadata{
					Pid:    data.Pid,
					Tgid:   data.Tgid,
					Uid:    data.Uid,
					Gid:    data.Gid,
					Binary: int8ArrayToString(data.Comm),
					Cwd:    cwd,
				},
			}

			return out, nil
		}
	}

	return kivev1alpha1.KiveAlert{}, fmt.Errorf("ReadAlert Error eBPF data received but no corresponsing kiveData was found")
}
