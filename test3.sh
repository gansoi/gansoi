#!/bin/sh

# Run this locally and visit https://gansoi-dev.com:9002/ in your favorite
# browser.

mkdir /tmp/gansoi-dev

cat >/tmp/gansoi-dev/node1.conf <<EOF
bind: "127.0.0.1:4934"
datadir: "/tmp/gansoi-dev/node1-data"

http:
  bind: "node1.gansoi-dev.com:9002"
  hostnames:
    - "gansoi-dev.com"
    - "node1.gansoi-dev.com"
  cert: "dockerroot/gansoi-dev.com-cert.pem"
  key: "dockerroot/gansoi-dev.com-key.pem"

redirect:
  bind: ""
EOF

cat >/tmp/gansoi-dev/node2.conf <<EOF
bind: "127.0.0.2:4934"
datadir: "/tmp/gansoi-dev/node2-data"

http:
  bind: "node2.gansoi-dev.com:9002"
  hostnames:
    - "gansoi-dev.com"
    - "node2.gansoi-dev.com"
  cert: "dockerroot/gansoi-dev.com-cert.pem"
  key: "dockerroot/gansoi-dev.com-key.pem"

redirect:
  bind: ""
EOF

cat >/tmp/gansoi-dev/node3.conf <<EOF
bind: "127.0.0.3:4934"
datadir: "/tmp/gansoi-dev/node3-data"

http:
  bind: "node3.gansoi-dev.com:9002"
  hostnames:
    - "gansoi-dev.com"
    - "node3.gansoi-dev.com"
  cert: "dockerroot/gansoi-dev.com-cert.pem"
  key: "dockerroot/gansoi-dev.com-key.pem"

redirect:
  bind: ""
EOF

mkdir /tmp/gansoi-dev/node1-data
mkdir /tmp/gansoi-dev/node2-data
mkdir /tmp/gansoi-dev/node3-data

./gansoi core --config /tmp/gansoi-dev/node1.conf init
TOKEN=$(./gansoi core --config /tmp/gansoi-dev/node1.conf print-token 2>/dev/null)

export DEBUG=*

# Start first server
tmux new-session -d './gansoi core --config /tmp/gansoi-dev/node1.conf run; sleep 10'

# Wait for the first node to self-elect
sleep 5

# Prepare the next two nodes.
./gansoi core --config /tmp/gansoi-dev/node2.conf join 127.0.0.1 $TOKEN
./gansoi core --config /tmp/gansoi-dev/node3.conf join 127.0.0.1 $TOKEN

# Start to nodes 3 seconds apart.
tmux split-window -v 'sleep 3 ; ./gansoi core --config /tmp/gansoi-dev/node2.conf run; sleep 10'
tmux split-window -v 'sleep 6 ; ./gansoi core --config /tmp/gansoi-dev/node3.conf run; sleep 10'

tmux -2 attach-session -d
