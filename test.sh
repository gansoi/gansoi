#!/bin/sh
tmux new-session -d './gansoi 127.0.0.1'
tmux split-window -v './gansoi 127.0.0.2'
tmux split-window -v './gansoi 127.0.0.3'
tmux -2 attach-session -d
