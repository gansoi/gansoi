#!/bin/sh

cat >/tmp/1.conf <<EOF
local = "local1.gansoi.com"
cert = "/etc/gansoi/cert.pem"
key = "/etc/gansoi/key.pem"
datadir = "/tmp/gansoi-1"
cluster = ["local1.gansoi.com", "local2.gansoi.com", "local3.gansoi.com"]
secret = "Qhzw5UTm66dWbHNyBdxazf6LfB7iaqcaRXzoot77gwMLErm1vb3VEDTBTupm8KxH"
EOF

cat >/tmp/2.conf <<EOF
local = "local2.gansoi.com"
cert = "/etc/gansoi/cert.pem"
key = "/etc/gansoi/key.pem"
datadir = "/tmp/gansoi-2"
cluster = ["local1.gansoi.com", "local2.gansoi.com", "local3.gansoi.com"]
secret = "Qhzw5UTm66dWbHNyBdxazf6LfB7iaqcaRXzoot77gwMLErm1vb3VEDTBTupm8KxH"
EOF

cat >/tmp/3.conf <<EOF
local = "local3.gansoi.com"
cert = "/etc/gansoi/cert.pem"
key = "/etc/gansoi/key.pem"
datadir = "/tmp/gansoi-3"
cluster = ["local1.gansoi.com", "local2.gansoi.com", "local3.gansoi.com"]
secret = "Qhzw5UTm66dWbHNyBdxazf6LfB7iaqcaRXzoot77gwMLErm1vb3VEDTBTupm8KxH"
EOF

mkdir -p /tmp/gansoi-1
mkdir -p /tmp/gansoi-2
mkdir -p /tmp/gansoi-3

sudo setcap CAP_NET_BIND_SERVICE=+eip ./gansoi

export DEBUG=*

tmux new-session -d './gansoi -config /tmp/1.conf'
tmux split-window -v './gansoi -config /tmp/2.conf'
tmux split-window -v './gansoi -config /tmp/3.conf'
tmux -2 attach-session -d
