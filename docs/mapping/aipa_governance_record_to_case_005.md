# AIPA Governance Record to DigiEmu Case 005 Mapping

Status: v0.18.1 boundary semantics cleanup  
Scope: Boundary mapping artifact  
Related DigiEmu case: Case 005 — Governance Record Continuity  

---

## 1. Purpose

This document maps a minimal AIPA Governance Record export structure to the DigiEmu Case 005 Governance Record Boundary model.

The purpose is not product integration.

The purpose is to clarify how a governance declaration artifact can remain separate from, but compatible with, independent deterministic continuity verification.

Core distinction:

```text
declared continuity != verified continuity
```

Related DigiEmu distinction:

```text
valid execution != coherent execution
```

This mapping is a boundary mapping, not a trust merger.

---

## 2. Boundary Model

The mapping preserves two independently evolvable layers.

### AIPA Layer

AIPA declares governance state:

- record identity
- schema version
- stage / status
- declared policy
- declared authority
- continuity mode
- review state
- approval state
- export metadata

AIPA answers:

```text
What was declared?
```

### DigiEmu Proof Layer

DigiEmu Proof evaluates verification anchors:

- policy fingerprint
- dependency fingerprint
- authority anchor
- previous baseline
- current values
- PASS / FAIL outcome
- issue list

DigiEmu Proof answers:

```text
Does the declared continuity verify?
```

---

## 3. Minimal Source Artifact

The sample source artifact is:

```text
docs/examples/aipa_governance_record_sample_001.json
```

It contains a minimal AIPA-like governance record with:

- `record_id`
- `record_type`
- `schema_version`
- `stage`
- `status`
- `declared_policy`
- `declared_authority`
- `continuity`
- `review_state`
- `export_metadata`

---

## 4. Minimal Target / Mapping Artifact

The DigiEmu-side mapping artifact is:

```text
docs/examples/digiemu_case_005_mapping_sample_001.json
```

It separates:

1. AIPA declaration fields
2. DigiEmu verification inputs
3. expected DigiEmu outcome
4. field boundary notes

---

## 5. Field Mapping

| AIPA Field | DigiEmu Case 005 Role | Boundary Meaning |
|---|---|---|
| `record_id` | record identity / baseline reference | Shared boundary identifier |
| `record_type` | declaration metadata | AIPA declaration field |
| `schema_version` | export/version metadata | Shared metadata, not proof result |
| `stage` | governance stage | AIPA declaration field |
| `status` | governance status | AIPA declaration field |
| `declared_policy.policy_ref` | declared policy reference | Shared boundary field |
| `declared_policy.policy_version` | policy metadata | AIPA declaration field unless fingerprinted |
| `declared_policy.policy_scope` | policy context | Declaration context |
| `declared_authority.authority_id` | declared authority | Shared boundary field |
| `declared_authority.authority_role` | authority metadata | AIPA declaration field |
| `declared_authority.authority_context` | declared authority context | AIPA declaration field; DigiEmu verifies the corresponding authority anchor |
| `continuity.continuity_mode` | declared continuity mode | Shared boundary field |
| `continuity.previous_record_id` | previous baseline reference | Shared boundary field |
| `continuity.declared_continuity_claim` | governance claim | AIPA declaration field |
| `review_state.review_status` | review declaration | AIPA declaration field |
| `review_state.approval_status` | approval declaration | AIPA declaration field |
| `review_state.reviewer` | reviewer metadata | AIPA declaration field |
| `review_state.approval_timestamp` | approval metadata | Audit metadata outside deterministic proof |
| `export_metadata.export_id` | export identity | AIPA/export metadata |
| `export_metadata.exported_at` | export metadata | Audit metadata outside verification hash |
| `export_metadata.export_format` | export metadata | AIPA/export metadata |
| `export_metadata.source_system` | source metadata | Audit metadata outside verification hash |

---

## 6. Verification Anchors

The AIPA Governance Record declares continuity.

DigiEmu Proof requires anchors to test that declaration.

Example anchors:

| DigiEmu Anchor | Purpose |
|---|---|
| `policy_fingerprint` | Verifies that the policy state did not silently mutate |
| `dependency_fingerprint` | Verifies that policy or workflow dependencies did not silently mutate |
| `authority_anchor` | Verifies that the declared authority context corresponds to a stable verification anchor |
| `previous baseline` | Defines the inherited continuity baseline |
| `current values` | Defines the state being checked against the baseline |

These are not merely governance claims.

They are verification evidence.

---

## 7. Authority Context vs Authority Anchor

AIPA may declare a human-readable `authority_context`.

DigiEmu Proof should not verify the human-readable authority context directly.

Instead:

```text
AIPA declares authority context.
DigiEmu verifies authority anchor.
```

This preserves the responsibility boundary.

AIPA remains responsible for governance semantics and human-readable authority declarations.

DigiEmu Proof remains responsible for verifying whether the corresponding deterministic anchor remains stable across the declared continuity boundary.

---

## 8. Audit Metadata vs Verification Inputs

Fields such as:

- `exported_at`
- `source_system`
- `approval_timestamp`
- `export_format`

are useful audit metadata.

They should remain outside deterministic verification unless a future version explicitly normalizes them into a proof boundary.

By default:

```text
audit metadata != verification input
```

This prevents contextual export information from accidentally becoming part of deterministic continuity validation.

---

## 9. Intentionally Separate Fields

The mapping intentionally avoids collapsing AIPA and DigiEmu Proof into one trust surface.

### AIPA review / approval state

Review and approval state remain governance declarations.

They should not be treated as verification results.

### DigiEmu PASS / FAIL

PASS / FAIL remains independently derived by DigiEmu Proof.

It should not be declared by AIPA.

### Fingerprints and anchors

Fingerprints and anchors remain verification evidence.

They should not be treated as governance claims.

### Export metadata

Export timestamps and source metadata are useful for audit context.

They should remain outside deterministic proof unless explicitly normalized and included in a verification boundary.

---

## 10. Example Failure

The sample mapping demonstrates this situation:

```text
AIPA declaration:
policy unchanged
authority unchanged
continuity mode = inherit
review status = reviewed
approval status = approved

DigiEmu anchors:
policy fingerprint unchanged
authority anchor unchanged
dependency fingerprint changed
```

Expected DigiEmu result:

```text
FAIL
```

Issue:

```text
dependency_fingerprint drift on inherit
```

Meaning:

```text
The governance record declares inherited continuity,
but DigiEmu Proof falsifies that continuity because a dependency anchor changed without explicit override.
```

---

## 11. Architectural Meaning

This mapping supports the shared boundary:

```text
AIPA says what was declared.
DigiEmu Proof says whether that declaration still holds.
```

This enables a reviewer to inspect both:

1. the declared governance record
2. the independently derived verification outcome

without requiring either layer to blindly trust the other.

This is a boundary mapping, not a trust merger.

---

## 12. Future Proof-Envelope Boundary

This mapping does not yet define a production integration.

However, it suggests a possible future boundary:

```text
AIPA Governance Record Export
        ->
DigiEmu Verification Input
        ->
Proof Envelope / Verification Outcome
```

In that future model:

- AIPA remains the governance declaration and workflow layer
- DigiEmu Proof remains the deterministic verification layer
- the exported artifact may become an input to a proof envelope
- ownership and trust surfaces remain distinct

The proof-envelope attachment point should occur after AIPA exports the declaration and before or inside DigiEmu verification.

AIPA should not be responsible for generating the DigiEmu PASS / FAIL result.

---

## 13. Summary

The minimal mapping demonstrates:

```text
declared continuity != verified continuity
```

A governance record can declare continuity, review, and approval.

DigiEmu Proof can still independently return FAIL if verification anchors drift without explicit override.

This keeps governance evidence and verification evidence related, but not identical.

The clean responsibility split is:

```text
AIPA declares policy reference.
DigiEmu verifies policy fingerprint.

AIPA declares authority context.
DigiEmu verifies authority anchor.

AIPA declares continuity mode.
DigiEmu evaluates baseline/current anchor continuity.

AIPA declares review/approval status.
DigiEmu independently returns PASS/FAIL.
```