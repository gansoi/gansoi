<!---
Sorry. This file is optimized for display on Github, not for viewing in
text-mode terminals.
-->

<h1 align="center">
  <br>
  <a href="http://gansoi.com/"><img src="https://gansoi.com/icon-transparent.svg" alt="Logo" width="250"></a>
  <br>
  Gansoi Infrastructure Monitoring
</h1>

<p align="center">
 <a href="https://travis-ci.org/gansoi/gansoi"><img src="https://travis-ci.org/gansoi/gansoi.svg?branch=master"></a>
 <a href="https://codebeat.co/projects/github-com-gansoi-gansoi"><img src="https://codebeat.co/badges/3f6e63cc-d32d-4962-a285-0f3a71d6cb80"></a>
<a href="https://app.fossa.io/projects/git%2Bgithub.com%2Fgansoi%2Fgansoi?ref=badge_shield" alt="FOSSA Status"><img src="https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgansoi%2Fgansoi.svg?type=shield"/></a>
 <a href="https://goreportcard.com/report/github.com/gansoi/gansoi"><img src="https://goreportcard.com/badge/github.com/gansoi/gansoi"></a>
</p>

Gansoi is a modern monitoring solution inspired by classical network monitoring
software (Nagios and friends) updated to modern and best practices.

### Disclaimer and Current State

At the moment this software is unusable for anyone but early adopters
and developers. It's a just-working proof-of-concept.

#### Current Limitations

- No useful agents implemented yet.
- The web interface is rudimentary at best. Useless at worst.

#### What We Currently Have

- A distributed core implemented using [Raft](https://raft.github.io/) and
  [BoltDB](https://github.com/boltdb/bolt).
- Full encryption of internal cluster communication using an internal
  [CA](https://en.wikipedia.org/wiki/Certificate_authority).
- The project requires nothing but a Go compiler when building. There
  is no runtime dependencies.
- A simple [Slack](https://slack.com/) notifier.
- A few example agents for testing.
- [LetsEncrypt](https://letsencrypt.org/) support for the web interface.
- It's quite performant. The smallest cloud server we could rent can easily
  handle hundreds of checks per second. With no optimizations yet.

### Mission Statement

Gansoi's mission is to radically improve self-hosted health monitoring and
alerting for ops and devops.

### Design goals

- [x] Fault tolerant and distributed - should tolerate multiple node failures
  with no single point of failure.
- [ ] Performant - should scale to thousands of checks per second on basic
  hardware.
- [x] Zero dependencies - should install and operate without any other
  software.
- [x] Geodistributed - should be deployable across the globe for full
  redundancy and geographically distributed checks.
- [ ] Multidimensional severity grouping - an alert should be categorized in
  both severity and urgency.
- [ ] No false positives - the system should never cry wolf.
- [x] Easy encryption - should support Let’s Encrypt out of the box.
- [x] Future-proof - Everything must support IPv6 out of the box.
- [x] Binary outcome - either a human should do something or not.

#### Alert integrations

- [x] Slack - it's what we all use and love.
- [ ] Pagerduty - oldtimers swear by this. Let's support it.
- [x] Email - CEO's want their alerts too.
- [ ] Twilio voice calls - for waking up your ops team at 3 AM.

#### Transports

Checks could be self-contained and run on a Gansoi node - or they can run on a
third party host. This will require a transport.

- [x] ssh - Gansoi should support SSH as a transport. It is universally supported
  as a remote access protocol. Gansoi must support some form of keep-alive to
  avoid the constant reconnecting and handshake.
- [ ] NRPE - This is the *industry standard* and we should support it.

### Building and development

Go 1.8 or newer is required for building Gansoi.

`go get ./...` should get all dependencies.

`go build .` should be enough to build Gansoi.

You can run a small local Gansoi cluster for testing and development:

    $ ./test3.sh

A three-node local cluster will be started shortly. You can visit the web
interface at https://gansoi-dev.com:9002/.

#### Docker

We provide a Docker image for your convenience.

##### Plain Docker

You can start a Gansoi node using regular Docker with a simple command:
`docker run --rm -p 80:80 -p 443:443 gansoi/gansoi`.

##### Compose and Stack

Included in this project is a `docker-compose.yml` for running Gansoi with
Docker Compose or Docker Stack.

If you're using Docker Compose you can try Gansoi using the command
`docker-compose up` and visiting https://gansoi-dev.com/.

The `docker-compose.yml` file can also be used for a Docker Stack deploy on
a Swarm: `docker stack deploy -c docker-compose.yml gansoi`

#### Slack Channel

If you would like you can join other developers in the
[#gansoi](https://gophers.slack.com/messages/gansoi/) channel on the Gophers
Slack (invites to Gophers Slack are available
[here](https://invite.slack.golangbridge.org/)).


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgansoi%2Fgansoi.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgansoi%2Fgansoi?ref=badge_large)