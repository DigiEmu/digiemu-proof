# DigiEmu Proof

Minimal prototype for deterministic execution verification.

---

## What this demonstrates

This prototype shows that a small execution path can be:

- captured as a deterministic state snapshot  
- serialized consistently  
- hashed with SHA-256  
- replayed  
- independently verified with PASS / FAIL  

---

## Core idea

```text
same input → same reconstructed state → same hash
Execution Flow
Intent envelope
→ Policy check
→ Deterministic action
→ Execution receipt
→ Canonical snapshot
→ SHA-256 hash
→ Replay
→ Verification result (PASS / FAIL)
Quick test (2 minutes)
git clone https://github.com/DigiEmu/digiemu-proof.git
cd digiemu-proof

go test ./...

go run ./cmd/digiemu-proof run input.json

go run ./cmd/digiemu-proof verify input.json sha256:9cde21367e09774964536e53d1e446c6ab9d18a6e36a964a60c6238f35463df4
Expected result
{
  "status": "PASS",
  "expected_hash": "sha256:9cde21367e09774964536e53d1e446c6ab9d18a6e36a964a60c6238f35463df4",
  "actual_hash": "sha256:9cde21367e09774964536e53d1e446c6ab9d18a6e36a964a60c6238f35463df4",
  "match": true
}
Determinism constraints
No timestamps
No randomness
No hidden state
No nondeterministic LLM output in the proof path

Everything that affects the hash must be reproducible.

Boundary principle
Inside hash:
deterministic, replayable state only

Outside hash:
metadata, timestamps, human context
Purpose

This prototype is a first step toward verifiable AI governance:

execution → receipt → state → replay → verification

The goal is to demonstrate that a small agent execution path can be:

reconstructed
independently verified
trusted without relying on runtime behavior
Versions
v0.1.0 — minimal deterministic execution proof
v0.2 — receipt + policy boundary model (current stable)