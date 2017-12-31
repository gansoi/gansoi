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

hosts:

  frankfurt.gansoi.com:
  london.gansoi.com:
  toronto.gansoi.com:

checks:

  gansoi http:
    agent: "http"
    arguments:
      url: "https://gansoi.com/"
    expressions:
      - "StatusCode == 200"
      - "SSLValidDays > 7"

  cluster node:
    agent: "process"
    arguments:
      name: "gansoi"
    hosts:
      - "frankfurt.gansoi.com"
      - "london.gansoi.com"
      - "toronto.gansoi.com"

EOF

export DEBUG=*

./gansoi core --config "${TARGET}/gansoi.yml" init
./gansoi core --config "${TARGET}/gansoi.yml" run
