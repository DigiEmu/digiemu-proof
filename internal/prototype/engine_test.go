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
