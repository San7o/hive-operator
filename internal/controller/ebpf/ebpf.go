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

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	container "github.com/San7o/hive-operator/internal/controller/container"
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

func ReadAlert(ctx context.Context, cli client.Reader) (hivev1alpha1.HiveAlert, error) {

	if RingbuffReader == nil {
		return hivev1alpha1.HiveAlert{}, fmt.Errorf("ReadAlert Error Ringbuffer not inizialized")
	}

	log := log.FromContext(ctx)
	data, err := ReadEbpfData() // Hangs
	if err != nil {
		return hivev1alpha1.HiveAlert{}, fmt.Errorf("ReadAlert Error Reading Ebpf Data: %w", err)
	}

	hiveDataList := &hivev1alpha1.HiveDataList{}
	err = cli.List(ctx, hiveDataList)
	if err != nil {
		return hivev1alpha1.HiveAlert{}, fmt.Errorf("ReadAlert Error Failed to get Hive Data resource: %w", err)
	}

	for _, hiveData := range hiveDataList.Items {
		if hiveData.Spec.InodeNo == data.Ino {

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
				log.Info(fmt.Sprintf("Could not read %s/%d/cwd while generating an HiveAlert, this can happen if the process terminated too quickly for the operator to react or the node is running in a container and procfs is not mounted in %s", container.ProcMountpoint, data.Pid, container.RealHostProcMountpoint))
			}
			out := hivev1alpha1.HiveAlert{
				Timestamp:      time.Now().Format(time.RFC3339),
				HivePolicyName: hiveData.Annotations["hive_policy_name"],
				Metadata: hivev1alpha1.HiveAlertMetadata{
					Path:     hiveData.Annotations["path"],
					Inode:    data.Ino,
					Mask:     data.Mask,
					KernelID: hiveData.Spec.KernelID,
					Callback: hiveData.ObjectMeta.Annotations["callback"],
				},
				Pod: hivev1alpha1.PodMetadata{
					Name:      hiveData.Annotations["pod_name"],
					Namespace: hiveData.Annotations["namespace"],
					Container: hivev1alpha1.ContainerMetadata{
						Id:   hiveData.Annotations["container_id"],
						Name: hiveData.Annotations["container_name"],
					},
					Ip: hiveData.Annotations["ip"],
				},
				Node: hivev1alpha1.NodeMetadata{
					Name: hiveData.Annotations["node_name"],
				},
				Process: hivev1alpha1.ProcessMetadata{
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

	return hivev1alpha1.HiveAlert{}, fmt.Errorf("ReadAlert Error eBPF data received but no corresponsing hiveData was found")
}
