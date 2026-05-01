# DigiEmu Proof

Minimal prototype for deterministic execution verification.

## What this demonstrates

This prototype shows that an execution path can be:

- captured as a deterministic state snapshot
- serialized consistently
- hashed with SHA-256
- replayed
- independently verified with PASS / FAIL
- validated for reference integrity

## Core idea

```text
same input → same execution path → same state → same hash
Flow
Intent envelope
→ Policy check
→ Deterministic action
→ Multi-step execution receipts
→ Canonical snapshot
→ SHA-256 hash
→ Replay
→ Verification result
→ Reference integrity validation
Quick test (2 minutes)
git clone https://github.com/DigiEmu/digiemu-proof.git
cd digiemu-proof

go test ./...

# ALLOW case
go run ./cmd/digiemu-proof run input.json
go run ./cmd/digiemu-proof verify input.json sha256:753149177e0016f5acdec2db44bb31f697899ae59155694eb6ee4a3bab3f1c4e

# DENY case
go run ./cmd/digiemu-proof run input-deny.json
go run ./cmd/digiemu-proof verify input-deny.json sha256:bddf3363fc38e941e70547ff538fcceebc01d8a407a9b8e6210bd91d50075a9c

# Reference validation
go run ./cmd/digiemu-proof validate input.json

Expected validation result:

{
  "status": "PASS",
  "issues": []
}
Determinism constraints

No timestamps
No randomness
No hidden state
No nondeterministic LLM output in the proof path

Everything that affects the hash must be reproducible.

Boundary rule
Inside hash:
- input
- policy result
- action result
- execution receipts

Outside hash:
- metadata (timestamps, notes, reviewers, environment)
Purpose

This prototype is a first step toward verifiable AI governance:

execution → receipt → state → replay → verification → validation

The goal is to show that an execution path can be reconstructed,
validated, and independently verified.

Versions
v0.1.0 — minimal deterministic execution proof
v0.2.0 — receipt + policy boundary model
v0.3.0 — metadata outside hash boundary
v0.4.0 — multi-step execution receipt chain
v0.5.0 — reference integrity validation
