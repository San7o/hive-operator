#!/bin/sh

source ./config.sh

tmux new-session -d -s $SESSION_NAME
tmux send-keys -t $SESSION_NAME "./scripts/master-node.sh" C-m
tmux split-window -v -t $SESSION_NAME
tmux send-keys -t $SESSION_NAME:0.1 "./scripts/worker1-node.sh" C-m
tmux attach -t $SESSION_NAME
