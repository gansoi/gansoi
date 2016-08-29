gansoi *
======

This software is unusable for now. It's a work in progress monitoring solution
for a modern world. As of the moment it's at most a non-working proof-of-concept.

Design goals
------------

- Fully distributed - no single point of failure.
- Fault tolerant - should tolerate multiple node failures.
- Performant - should scale to thousands of checks per second on basic hardware.
- Zero dependencies - should install and operate without any other software.
- Geodistributed - should be deployable across the globe for full redundancy and geographically distributed checks.
- Multidimensional severity grouping - an alert should be categorized in both severity and urgency.
- No false positives - the system should never cry wolf.
- Easy encryption - should support Let’s Encrypt out of the box.
- Future-proof - Everything must support IPv6 out of the box.

\* gansoi is a working title.
