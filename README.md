# DigiEmu Proof

Minimal prototype for deterministic execution verification.

## What this demonstrates

This prototype shows that a small execution path can be:

- captured as a deterministic state snapshot
- serialized consistently
- hashed with SHA-256
- replayed
- independently verified with PASS / FAIL

## Core idea

```text
same input → same reconstructed state → same hash
