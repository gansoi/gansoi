#!/bin/bash

TARGET="${TMPDIR:-/tmp/}/gansoi-dev"

mkdir -p "${TARGET}"

cat >"${TARGET}/gansoi.yml" <<EOF
bind: "127.0.0.1:4934"
datadir: "${TARGET}"

http:
  bind: "node1.gansoi-dev.com:9002"
  hostnames:
    - "gansoi-dev.com"
    - "node1.gansoi-dev.com"
  cert: "dockerroot/gansoi-dev.com-cert.pem"
  key: "dockerroot/gansoi-dev.com-key.pem"
EOF

export DEBUG=*

./gansoi core --config "${TARGET}/gansoi.yml" init
./gansoi core --config "${TARGET}/gansoi.yml" run
