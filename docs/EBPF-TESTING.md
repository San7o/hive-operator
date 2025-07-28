# eBPF Testing

In this document we discuss how to test the eBPF program. We assume
that you have the required dependencies, if not please check the
[setup](./SETUP.md) document so that you are able to compile the
program.

The code of the eBPF program lives in the `ebpf/` directory of the
project. A go loader is implemented in `test/ebpf-local/` which will
load the ebpf program on your machine. To help you during development,
the Makefile at the root directory comes with the following helpful
make commands:

- `test-build-ebpf`: build the ebpf-local program
- `test-run-ebpf`: run the ebpf-local program
- `test-ebpf`: build and run the ebpf-local program

You often need some kind of server to send network traffic to. You can
use the `callback/` server which will echo the requests they receive
and is accessible via `http://localhost:8090/callback`.

To run both the eBPF program and the callback service, you can create
a tmux session with:

```
make test-tmux
```

This will create a session with three panes split horizontally: one
with the eBPF program, one with the callback server, and one with a
bash shell.
