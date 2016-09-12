gansoi *
======

This software is unusable for now. It's a work in progress monitoring solution
for a modern world. As of the moment it's at most a non-working proof-of-concept.

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

\* gansoi is a working title.
