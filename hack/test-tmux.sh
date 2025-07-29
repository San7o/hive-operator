#!/bin/sh

#  This script creates a tmux session with two panes split
#  horizontally: the first pane will run the ebpf program locally,
#  which will print the eBPF data on his standard output; the second
#  pane will run bash so you can send commands
#
#  This script should be executed from the root directory of the
#  project. There, you can run this script via
#
#      $ make test-tmux
#
#  Happy hacking!

SESSION_NAME=test-ebpf-local

tmux new-session -d -s $SESSION_NAME
tmux send-keys -t $SESSION_NAME "make test-ebpf" C-m
tmux split-window -v -t $SESSION_NAME
tmux send-keys -t $SESSION_NAME:0.1 "/bin/bash" C-m
tmux attach -t $SESSION_NAME
