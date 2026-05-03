# DigiEmu Proof

Minimal prototype for deterministic execution and transition verification.

---

## What this demonstrates

This prototype shows that an execution path can be:

* captured as deterministic state snapshots  
* serialized consistently  
* hashed with SHA-256  
* replayed  
* independently verified with PASS / FAIL  

Now extended to:

* deterministic transition verification  
* full chain continuity validation  
* reproducible failure semantics  

---

## Core idea

```text
same input → same reconstructed state → same hash
```
Extended to:

```text
state₀ → transition → state₁ → ... → stateₙ (verifiable)
```
## Execution Flow

```text
Intent envelope
→ Policy check
→ Deterministic action
→ Transition receipt
→ Canonical state
→ SHA-256 hash
→ Replay
→ Verification (PASS / FAIL)
```
## Transition Proof (v0.8)

Each transition is verified using a receipt that binds:

prev_state_hash → receipt → next_state_hash

And proves:

- intent_id
- policy_id
- policy_decision
- action_id
- action_type
```md
- input_ref / output_ref
```

This ensures:

- every transition is deterministic and verifiable

---

## Decision Proof Surface (v0.9)

v0.9 introduces a minimal governance verification surface.

Execution integrity proves:

```text


Did the system do what the receipt says it did?

## Chain Integrity (v0.7)

Full chain verification enforces:

len(receipts) == len(states) - 1

and:

receipt[i].prev_state_hash == hash(states[i])
receipt[i].next_state_hash == hash(states[i+1])

This guarantees:

- no missing transitions
- no reordered steps
- no hidden state gaps


## Failure Semantics
valid execution → PASS  
invalid execution → FAIL (reproducible)

This means:

failure is part of the proof surface
## Quick test (2 minutes)


```bash
git clone https://github.com/DigiEmu/digiemu-proof.git
cd digiemu-proof

go test ./...

go run ./cmd/digiemu-proof run input.json

go run ./cmd/digiemu-proof verify input.json sha256:9cde21367e09774964536e53d1e446c6ab9d18a6e36a964a60c6238f35463df4
```


## Expected result

```json
{
  "status": "PASS",
  "expected_hash": "...",
  "actual_hash": "...",
  "match": true
}
```

## Determinism constraints

- No timestamps
- No randomness
- No hidden state
- No nondeterministic LLM output in the proof path

Everything that affects the hash must be reproducible.

## Boundary principle

Inside hash:
deterministic, replayable state

Outside hash:
metadata, timestamps, environment, human context

## Purpose

This prototype demonstrates:

state → transition → state → replay → verification

The system proves:

deterministic state reconstruction
deterministic transitions
deterministic failure behavior

## Versions

* v0.1.0 — deterministic execution proof
* v0.2 — boundary model (policy / receipt separation)
* v0.6.1 — transition verification
* v0.6.2 — independent next state derivation
* v0.6.3 — strict verification semantics
* v0.7.0 — chain continuity verification
* v0.7.1 — hardened failure semantics + negative tests
* v0.8.0 — transition receipt proof fields
* v0.9.0 — decision proof surface / governance-bound receipt