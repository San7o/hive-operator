package ebpf

// To interact with the BPF program, we can (and should) generate
// some go bindings using the following command:
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type log_data -tags linux -cflags "-D __${ARCH}__" bpf ../../../ebpf/tracer.bpf.c
