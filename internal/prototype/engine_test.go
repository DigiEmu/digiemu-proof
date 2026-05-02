package prototype

import "testing"

func TestDeterministicSnapshotHash(t *testing.T) {
	input := IntentEnvelope{
		Intent: "summarize_text",
		Context: map[string]string{
			"text": "DigiEmu Core verifies deterministic knowledge states.",
		},
	}

	snapshot1 := BuildSnapshot(input)
	hash1, err := HashSnapshot(snapshot1)
	if err != nil {
		t.Fatal(err)
	}

	snapshot2 := BuildSnapshot(input)
	hash2, err := HashSnapshot(snapshot2)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Fatalf("expected identical hashes, got %s and %s", hash1, hash2)
	}
}

func TestPolicyAllowWhenTextPresent(t *testing.T) {
	input := IntentEnvelope{
		Intent: "summarize_text",
		Context: map[string]string{
			"text": "DigiEmu Core verifies deterministic knowledge states.",
		},
	}

	policy := EvaluatePolicy(input)

	if policy.Decision != "allow" {
		t.Fatalf("expected allow, got %s", policy.Decision)
	}

	if policy.ReasonCode != "TEXT_PRESENT" {
		t.Fatalf("expected TEXT_PRESENT, got %s", policy.ReasonCode)
	}
}

func TestPolicyDenyWhenTextMissing(t *testing.T) {
	input := IntentEnvelope{
		Intent:  "summarize_text",
		Context: map[string]string{},
	}

	policy := EvaluatePolicy(input)

	if policy.Decision != "deny" {
		t.Fatalf("expected deny, got %s", policy.Decision)
	}

	if policy.ReasonCode != "TEXT_MISSING_OR_UNSUPPORTED_INTENT" {
		t.Fatalf("unexpected reason code: %s", policy.ReasonCode)
	}
}

func TestVerifyPass(t *testing.T) {
	input := IntentEnvelope{
		Intent: "summarize_text",
		Context: map[string]string{
			"text": "DigiEmu Core verifies deterministic knowledge states.",
		},
	}

	snapshot := BuildSnapshot(input)
	expectedHash, err := HashSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Verify(input, expectedHash)
	if err != nil {
		t.Fatal(err)
	}

	if result.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Status)
	}

	if !result.Match {
		t.Fatal("expected hash match")
	}
}

func TestTransitionV06CreatesVisibleStateChange(t *testing.T) {
	state0 := CanonicalStateV06{
		SchemaVersion: "canonical-state.v0.6",
		Intent: IntentRef{
			ID:       "intent.summary.v1",
			InputRef: "input.text.v1",
		},
		Policy: PolicyRef{
			ID:       "policy.allow_summary.v1",
			Decision: "allow",
		},
		Action: ActionRef{
			ID:        "action.summary.v1",
			Type:      "summary",
			OutputRef: "output.summary.v1",
		},
		Refs: map[string]string{
			"input.text.v1":           HashStringV06("DigiEmu Core verifies deterministic knowledge states."),
			"policy.allow_summary.v1": HashStringV06("intent == summarize_text && context.text != empty"),
		},
	}

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	receipt, state1, err := BuildTransitionV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	if prevHash == nextHash {
		t.Fatal("expected state transition to change the canonical state hash")
	}

	if state1.Refs["output.summary.v1"] == "" {
		t.Fatal("expected output.summary.v1 ref to be created")
	}

	if receipt.PrevStateHash != prevHash {
		t.Fatalf("expected prev_state_hash %s, got %s", prevHash, receipt.PrevStateHash)
	}

	if receipt.NextStateHash != nextHash {
		t.Fatalf("expected next_state_hash %s, got %s", nextHash, receipt.NextStateHash)
	}
}

func TestVerifyTransitionV06FailsOnTamperedNextState(t *testing.T) {
	state0 := CanonicalStateV06{
		SchemaVersion: "canonical-state.v0.6",
		Intent: IntentRef{
			ID:       "intent.summary.v1",
			InputRef: "input.text.v1",
		},
		Policy: PolicyRef{
			ID:       "policy.allow_summary.v1",
			Decision: "allow",
		},
		Action: ActionRef{
			ID:        "action.summary.v1",
			Type:      "summary",
			OutputRef: "output.summary.v1",
		},
		Refs: map[string]string{
			"input.text.v1":           HashStringV06("DigiEmu Core verifies deterministic knowledge states."),
			"policy.allow_summary.v1": HashStringV06("intent == summarize_text && context.text != empty"),
		},
	}

	receipt, state1, err := BuildTransitionV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	state1.Refs["output.summary.v1"] = HashStringV06("tampered output")

	result, err := VerifyTransitionV06(state0, receipt, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if result.Match {
		t.Fatal("expected transition mismatch")
	}
}
