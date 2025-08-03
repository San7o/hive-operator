# eBPF Testing

In this document we discuss how to test the eBPF program. We assume
that you have the required dependencies, if not please check the
[DEVELOPMENT](./DEVELOPMENT.md) document so that you are able to
compile the program.

The code of the eBPF program lives in the `ebpf/` directory of the
project. A go loader is implemented in `test/ebpf-local/` which will
load the ebpf program on your machine. To help you during development,
the Makefile at the root directory comes with the following helpful
make commands:

- `test-build-ebpf`: build the ebpf-local program
- `test-run-ebpf`: run the ebpf-local program
- `test-ebpf`: build and run the ebpf-local program

If you want to receive all the alerts in a single place, you can use
setup a callback service from `callback/` by running `make docker` and
`make deploy`. Then, you can configure your policies to send a
callback to
`http://callback-service.hive-operator-system.svc.cluster.local:9376/ingest`.

For convenience, you can create a tmux session with:

```bash
make test-tmux
```

This will create a session with two panes split horizontally: one with
the eBPF program and one with a bash shell.
