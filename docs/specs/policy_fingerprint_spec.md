# Policy Fingerprint Specification

## Purpose

A policy fingerprint represents the effective policy context applied during a transition.

It is not only a hash of the policy name or policy reference.

It binds the declared policy to the context in which it was actually applied.

---

## Fingerprint Definition

```text
policy_fingerprint = hash(
  policy_definition
  + resolved_dependencies
  + authority_scope
  + execution_constraints
  + external_context_bindings
)
```

---

## Components

### policy_definition

The deterministic policy artifact or rule set used for the transition.

### resolved_dependencies

All dependency versions required to evaluate the policy.

### authority_scope

The governance authority under which the policy was applied.

### execution_constraints

The constraints active during execution.

### external_context_bindings

External references that affect policy interpretation.

---

## Validation Rule

```text
inherit  → policy_fingerprint must remain continuous
override → policy_fingerprint may change if explicitly declared
```

---

## Failure Modes

```text
missing fingerprint → FAIL
fingerprint mismatch on inherit → FAIL
undeclared dependency mutation → FAIL
declared override with invalid context → FAIL
```

---

## Design Principle

Canonical state remains reconstruction-focused.

Receipts carry transition intent and policy context evidence.

Validators compare:

- continuity  
- declared override  
- effective context alignment  