package main

import (
	"context"
	"os"

	controller "github.com/San7o/hive-operator/internal/controller"
	ebpf "github.com/San7o/hive-operator/internal/controller/ebpf"

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

	ctrl.SetLogger(zap.New())

	ctx := context.Background()
	log := log.FromContext(ctx)

	reconciler := controller.HivePolicyReconciler{}

	if err := ebpf.LoadEbpf(ctx); err != nil {
		log.Error(err, "Error loading eBPF program")
		return
	}
	defer ebpf.UnloadEbpf(ctx)

	log.Info("Logging data...")
	controller.Output(reconciler.UncachedClient)

	return
}
