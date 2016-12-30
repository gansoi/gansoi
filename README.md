gansoi
======

Gansoi is a modern monitoring solution inspired by classical network monitoring
software (Nagios and friends) updated to modern and best practices.

[![Build Status](https://travis-ci.org/gansoi/gansoi.svg?branch=master)](https://travis-ci.org/gansoi/gansoi)
[![codebeat badge](https://codebeat.co/badges/3f6e63cc-d32d-4962-a285-0f3a71d6cb80)](https://codebeat.co/projects/github-com-gansoi-gansoi)

### Disclaimer and Current State

At the momemt this software is unusable for anyone but *very* early adopters
and developers. It's a just-working proof-of-concept.

#### Current Limitations

- No useful agents implemented yet.
- The web interface is rudimentary at best. Useless at worst.
- Single node operation is not supported. A cluster of core nodes *is* needed.
  You can however run them on the same host.

#### What We Currently Have

- Distributed core is implemented using [Raft](https://raft.github.io/) and
  [BoltDB](https://github.com/boltdb/bolt).
- The project currently requires nothing but a Go compiler when building. There
  is no runtime dependencies.
- A very simple [Slack](https://slack.com/) notifier is implemented.
- A few example agents are implemented for testing.
- [LetsEncrypt](https://letsencrypt.org/) is supported for node-to-node
  communication and web interface.
- It's quite performant. The smallest cloud server we could rent can easily
  handle hundreds of checks per second. With no optimizations yet.

### Design goals

- Fault tolerant and distributed - should tolerate multiple node failures with no single point of failure.
- Performant - should scale to thousands of checks per second on basic hardware.
- Zero dependencies - should install and operate without any other software.
- Geodistributed - should be deployable across the globe for full redundancy and geographically distributed checks.
- Multidimensional severity grouping - an alert should be categorized in both severity and urgency.
- No false positives - the system should never cry wolf.
- Easy encryption - should support Let’s Encrypt out of the box.
- Future-proof - Everything must support IPv6 out of the box.

#### Alert integrations

- Slack - it's what we all use and love.
- Pagerduty - oldtimers swear by this. Let's support it.
- Email - CEO's want their alerts too.
- Twilio voice calls - for waking up your ops team at 3AM.

#### Transports

Checks could be selfcontained and run on a Gansoi node - or they can run on a
third party host. This will require a transport.

- ssh - Gansoi should support SSH as a transport. It is universally supported as a remote access protocol. Gansoi must support some form of keep-alive to avoid the constant reconnecting and handshake.
- NRPE - This is the *industry standard* and we should support it.

### Building and development

Go 1.6 or newer is required for building gansoi.

`go get ./...` should get all dependencies.

`go build .` should be enough to build gansoi.

You can run a small local gansoi cluster for testing and development:

    $ ./test.sh

A three-node local cluster should be started shortly. You can visit the webinterface at https://gansoi-dev.com:9002/.
