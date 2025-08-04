# eBPF Testing

This document contains information to build the eBPF program locally
without building the entire operator or running a kubernetes
cluster. We assume that you have the required dependencies, if not
please check the [DEVELOPMENT](./DEVELOPMENT.md) document so that you
are able to compile the program.

The code of the eBPF program lives in the `ebpf/` directory of the
project. A go loader is implemented in `test/ebpf-local/` which will
load the ebpf program on your machine. To help you during development,
the Makefile at the root directory comes with the following helpful
make commands:

- `test-build-ebpf`: build the ebpf-local program
- `test-run-ebpf`: run the ebpf-local program
- `test-ebpf`: build and run the ebpf-local program

By default, the program monitors accesses to the `LICENSE.md` file in
the root directory, so you can test that it is working as expected.

For convenience, you can create a tmux session with:

```bash
make test-tmux
```

This will create a session with two panes split horizontally: one with
the eBPF program and one with a bash shell.
