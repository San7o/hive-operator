package controller

const (
	// The name used by our controller to claim ownership of fields when doing server-side apply in Kubernetes.
	FieldOwnerKiveController = "kive-controller"

	// Where to find the identifier of this running kernel
	KernelIDPath = "/proc/sys/kernel/random/boot_id"

	// Label used for the trap identifier
	TrapIdLabel = "trap-id"
)
