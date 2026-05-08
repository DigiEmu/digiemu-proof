\# AIPA Governance Record to DigiEmu Case 005 Mapping



Status: v0.18.0 mapping draft  

Scope: Boundary mapping artifact  

Related DigiEmu case: Case 005 — Governance Record Continuity  



\---



\## 1. Purpose



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



\---



\## 2. Boundary Model



The mapping preserves two independently evolvable layers.



\### AIPA Layer



AIPA declares governance state:



\- record identity

\- schema version

\- stage / status

\- declared policy

\- declared authority

\- continuity mode

\- review state

\- approval state

\- export metadata



AIPA answers:



```text

What was declared?

```



\### DigiEmu Proof Layer



DigiEmu Proof evaluates verification anchors:



\- policy fingerprint

\- dependency fingerprint

\- authority anchor

\- previous baseline

\- current values

\- PASS / FAIL outcome

\- issue list



DigiEmu Proof answers:



```text

Does the declared continuity verify?

```



\---



\## 3. Minimal Source Artifact



The sample source artifact is:



```text

docs/examples/aipa\_governance\_record\_sample\_001.json

```



It contains a minimal AIPA-like governance record with:



\- `record\_id`

\- `record\_type`

\- `schema\_version`

\- `stage`

\- `status`

\- `declared\_policy`

\- `declared\_authority`

\- `continuity`

\- `review\_state`

\- `export\_metadata`



\---



\## 4. Minimal Target / Mapping Artifact



The DigiEmu-side mapping artifact is:



```text

docs/examples/digiemu\_case\_005\_mapping\_sample\_001.json

```



It separates:



1\. AIPA declaration fields

2\. DigiEmu verification inputs

3\. expected DigiEmu outcome

4\. field boundary notes



\---



\## 5. Field Mapping



| AIPA Field | DigiEmu Case 005 Role | Boundary Meaning |

|---|---|---|

| `record\_id` | record identity / baseline reference | Shared boundary identifier |

| `record\_type` | declaration metadata | AIPA declaration field |

| `schema\_version` | export/version metadata | Shared metadata, not proof result |

| `stage` | governance stage | AIPA declaration field |

| `status` | governance status | AIPA declaration field |

| `declared\_policy.policy\_ref` | declared policy reference | Shared boundary field |

| `declared\_policy.policy\_version` | policy metadata | AIPA declaration field unless fingerprinted |

| `declared\_policy.policy\_scope` | policy context | Declaration context |

| `declared\_authority.authority\_id` | declared authority | Shared boundary field |

| `declared\_authority.authority\_role` | authority metadata | AIPA declaration field |

| `declared\_authority.authority\_context` | authority context | Shared boundary field |

| `continuity.continuity\_mode` | declared continuity mode | Shared boundary field |

| `continuity.previous\_record\_id` | previous baseline reference | Shared boundary field |

| `continuity.declared\_continuity\_claim` | governance claim | AIPA declaration field |

| `review\_state.review\_status` | review declaration | AIPA declaration field |

| `review\_state.approval\_status` | approval declaration | AIPA declaration field |

| `review\_state.reviewer` | reviewer metadata | AIPA declaration field |

| `review\_state.approval\_timestamp` | approval metadata | Metadata outside deterministic proof |

| `export\_metadata.export\_id` | export identity | AIPA/export metadata |

| `export\_metadata.exported\_at` | export metadata | Outside verification hash |

| `export\_metadata.export\_format` | export metadata | AIPA/export metadata |

| `export\_metadata.source\_system` | source metadata | AIPA/export metadata |



\---



\## 6. Verification Anchors



The AIPA Governance Record declares continuity.



DigiEmu Proof requires anchors to test that declaration.



Example anchors:



| DigiEmu Anchor | Purpose |

|---|---|

| `policy\_fingerprint` | Verifies that the policy state did not silently mutate |

| `dependency\_fingerprint` | Verifies that policy or workflow dependencies did not silently mutate |

| `authority\_anchor` | Verifies that the authority context remained stable |

| `previous baseline` | Defines the inherited continuity baseline |

| `current values` | Defines the state being checked against the baseline |



These are not merely governance claims.



They are verification evidence.



\---



\## 7. Intentionally Separate Fields



The mapping intentionally avoids collapsing AIPA and DigiEmu Proof into one trust surface.



\### AIPA review / approval state



Review and approval state remain governance declarations.



They should not be treated as verification results.



\### DigiEmu PASS / FAIL



PASS / FAIL remains independently derived by DigiEmu Proof.



It should not be declared by AIPA.



\### Fingerprints and anchors



Fingerprints and anchors remain verification evidence.



They should not be treated as governance claims.



\### Export metadata



Export timestamps and source metadata are useful for audit context.



They should remain outside deterministic proof unless explicitly normalized and included in a verification boundary.



\---



\## 8. Example Failure



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

dependency\_fingerprint drift on inherit

```



Meaning:



```text

The governance record declares inherited continuity,

but DigiEmu Proof falsifies that continuity because a dependency anchor changed without explicit override.

```



\---



\## 9. Architectural Meaning



This mapping supports the shared boundary:



```text

AIPA says what was declared.

DigiEmu Proof says whether that declaration still holds.

```



This enables a reviewer to inspect both:



1\. the declared governance record

2\. the independently derived verification outcome



without requiring either layer to blindly trust the other.



\---



\## 10. Future Proof-Envelope Boundary



This mapping does not yet define a production integration.



However, it suggests a possible future boundary:



```text

AIPA Governance Record Export

&#x20;       ->

DigiEmu Verification Input

&#x20;       ->

Proof Envelope / Verification Outcome

```



In that future model:



\- AIPA remains the governance declaration and workflow layer

\- DigiEmu Proof remains the deterministic verification layer

\- the exported artifact may become an input to a proof envelope

\- ownership and trust surfaces remain distinct



\---



\## 11. Summary



The minimal mapping demonstrates:



```text

declared continuity != verified continuity

```



A governance record can declare continuity, review, and approval.



DigiEmu Proof can still independently return FAIL if verification anchors drift without explicit override.



This keeps governance evidence and verification evidence related, but not identical.

