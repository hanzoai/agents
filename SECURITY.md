# Security Policy

## Reporting a vulnerability

Email security@hanzo.ai with details. Encrypt with our PGP key (fingerprint TBD).

We respond within 48 hours. Critical issues receive same-day acknowledgment.

## Scope

This policy covers code in this repository. For the broader Hanzo platform threat model, see [hanzoai/HIPs](https://github.com/hanzoai/HIPs).

## Sandbox boundary

Agents in this repo execute tool calls through `hanzo-mcp` and inherit its sandbox — every shell or filesystem action runs with the calling user's permissions on the user's own machine. Secret reads are gated by `kms` per-secret AI policy; outgoing network actions are subject to whatever egress policy the host environment configures.

For runtime sandbox guarantees, see HIP-0105 (in-process extension runtimes).
