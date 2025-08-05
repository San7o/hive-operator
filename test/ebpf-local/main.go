/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"context"
	"os"
	"strings"

	controller "github.com/San7o/kivebpf/internal/controller"
	ebpf "github.com/San7o/kivebpf/internal/controller/ebpf"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	InterfaceName = "lo"
	Port          = "8090"
)

func main() {

	if len(os.Args) > 1 {
		InterfaceName = os.Args[1]
	}

	opts := zap.Options{
		Development: true, // Enables console encoder and disables sampling
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := context.Background()
	log := log.FromContext(ctx)

	// Set the KernelID
	kernelIDBytes, err := os.ReadFile(controller.KernelIDPath)
	if err != nil {
		log.Error(err, "Cannot read kernel boot ID at "+controller.KernelIDPath)
	}
	controller.KernelID = string(kernelIDBytes)
	controller.KernelID = strings.TrimSpace(controller.KernelID)

	if err := ebpf.LoadEbpf(ctx); err != nil {
		log.Error(err, "Error loading eBPF program")
		return
	}
	defer ebpf.UnloadEbpf(ctx)

	ino, err := GetInode("LICENSE.md")
	if err != nil {
		log.Error(err, "Error Get Inode")
	}

	var key uint32 = 0
	err = ebpf.UpdateTracedInodes(key, ino)
	if err != nil {
		log.Error(err, "Error Update map")
	}

	log.Info("Logging data...")

	// Read and print loop
	for {
		data, err := ebpf.ReadEbpfData() // Hangs
		if err != nil {
			log.Error(err, "Error Read Ebpf data")
		}

		log.Info("Received Data", "pid", data.Pid, "inode", data.Ino)
	}
}
