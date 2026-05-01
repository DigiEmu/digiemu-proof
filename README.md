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
```

## Flow

```text
Intent envelope
→ Policy check
→ Deterministic action
→ Execution receipt
→ Canonical snapshot
→ SHA-256 hash
→ Replay
→ Verification result
```

## Quick test (2 minutes)

```bash
git clone https://github.com/DigiEmu/digiemu-proof.git
cd digiemu-proof
go test ./...
go run ./cmd/digiemu-proof run input.json
go run ./cmd/digiemu-proof verify input.json sha256:4ce910afcf002a5209ce6a1aed790e6238a0a6385867777e7ce3e26c651168ed
```

Expected result:

```json
{
  "status": "PASS",
  "expected_hash": "sha256:4ce910afcf002a5209ce6a1aed790e6238a0a6385867777e7ce3e26c651168ed",
  "actual_hash": "sha256:4ce910afcf002a5209ce6a1aed790e6238a0a6385867777e7ce3e26c651168ed",
  "match": true
}
```

## Determinism constraints

No timestamps  
No randomness  
No hidden state  
No nondeterministic LLM output in the proof path  

Everything that affects the hash must be reproducible.

## Purpose

This prototype is a first step toward verifiable AI governance:

```text
execution → receipt → state → replay → verification
```

The goal is to show that a small agent execution path can be reconstructed and independently verified.

## Versions

- `v0.1.0` — minimal deterministic execution proof  
- `v0.2` — minimal receipt + policy boundary model