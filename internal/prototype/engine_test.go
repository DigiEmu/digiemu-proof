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

func TestVerifyTransitionReceiptV08FailsOnPolicyDecisionMismatch(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v08"

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	receipt := TransitionReceiptV08{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: "deny",
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,
		Status:         "PASS",
	}

	result, err := VerifyTransitionReceiptV08(state0, receipt, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected mismatch")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "policy_decision mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyTransitionReceiptV08Pass(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v08"

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	receipt := TransitionReceiptV08{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: state0.Policy.Decision,
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,
		Status:         "PASS",
	}

	result, err := VerifyTransitionReceiptV08(state0, receipt, state1)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %+v", result.Issues)
	}

	if result.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Status)
	}

	if len(result.Issues) != 0 {
		t.Fatalf("expected no issues, got %+v", result.Issues)
	}
}

func TestVerifyTransitionReceiptV09Pass(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v09"

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	receipt := TransitionReceiptV09{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: state0.Policy.Decision,
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,

		PolicySetHash:         HashStringV06("policy-set:v09"),
		AuthorizationContext:  "authz-context:v09",
		ConstraintResult:      "allow",
		ActorID:               "agent:test-runner",
		CapabilityScope:       "summary:create",
		SequenceBoundary:      "seq:1",
		DependencyFingerprint: HashStringV06("dependency-state:v09"),

		Status: "PASS",
	}

	result, err := VerifyTransitionReceiptV09(state0, receipt, state1)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %+v", result.Issues)
	}

	if result.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Status)
	}

	if len(result.Issues) != 0 {
		t.Fatalf("expected no issues, got %+v", result.Issues)
	}
}

func TestVerifyTransitionReceiptV09FailsOnMissingPolicySetHash(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v09"

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	receipt := TransitionReceiptV09{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: state0.Policy.Decision,
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,

		PolicySetHash:         "",
		AuthorizationContext:  "authz-context:v09",
		ConstraintResult:      "allow",
		ActorID:               "agent:test-runner",
		CapabilityScope:       "summary:create",
		SequenceBoundary:      "seq:1",
		DependencyFingerprint: HashStringV06("dependency-state:v09"),

		Status: "PASS",
	}

	result, err := VerifyTransitionReceiptV09(state0, receipt, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on missing policy_set_hash")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "policy_set_hash missing" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyTransitionReceiptV09FailsOnDeniedConstraint(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v09"

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	receipt := TransitionReceiptV09{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: state0.Policy.Decision,
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,

		PolicySetHash:         HashStringV06("policy-set:v09"),
		AuthorizationContext:  "authz-context:v09",
		ConstraintResult:      "deny",
		ActorID:               "agent:test-runner",
		CapabilityScope:       "summary:create",
		SequenceBoundary:      "seq:1",
		DependencyFingerprint: HashStringV06("dependency-state:v09"),

		Status: "FAIL",
	}

	result, err := VerifyTransitionReceiptV09(state0, receipt, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on denied constraint")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "constraint_result not allow" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func buildV10ExecutionReceipt(t *testing.T, state0 CanonicalStateV06, state1 CanonicalStateV06) TransitionReceiptV08 {
	t.Helper()

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	return TransitionReceiptV08{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: state0.Policy.Decision,
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,
		Status:         "PASS",
	}
}

func buildV10DecisionReceipt(t *testing.T, state0 CanonicalStateV06, state1 CanonicalStateV06) TransitionReceiptV09 {
	t.Helper()

	prevHash, err := HashCanonicalStateV06(state0)
	if err != nil {
		t.Fatal(err)
	}

	nextHash, err := HashCanonicalStateV06(state1)
	if err != nil {
		t.Fatal(err)
	}

	return TransitionReceiptV09{
		StepID:         "step-1",
		Actor:          "test-runner",
		ActionType:     state0.Action.Type,
		IntentID:       state0.Intent.ID,
		PolicyID:       state0.Policy.ID,
		PolicyDecision: state0.Policy.Decision,
		ActionID:       state0.Action.ID,
		InputRef:       state0.Intent.InputRef,
		PolicyRef:      state0.Policy.ID,
		OutputRef:      state0.Action.OutputRef,
		PrevStateHash:  prevHash,
		NextStateHash:  nextHash,

		PolicySetHash:         HashStringV06("policy-set:v10"),
		AuthorizationContext:  "authz-context:v10",
		ConstraintResult:      "allow",
		ActorID:               "agent:test-runner",
		CapabilityScope:       "summary:create",
		SequenceBoundary:      "seq:1",
		DependencyFingerprint: HashStringV06("dependency-state:v10"),

		Status: "PASS",
	}
}

func TestVerifyProofEnvelopeV10Pass(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v10"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	envelope, err := BuildProofEnvelopeV10(execution, decision)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV10(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %+v", result.Issues)
	}

	if result.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
}

func TestVerifyProofEnvelopeV10FailsOnEnvelopeHashMismatch(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v10"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	envelope, err := BuildProofEnvelopeV10(execution, decision)
	if err != nil {
		t.Fatal(err)
	}

	envelope.EnvelopeHash = "sha256:tampered"

	result, err := VerifyProofEnvelopeV10(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on envelope hash mismatch")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "envelope_hash mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV10FailsOnDecisionProofInvalid(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v10"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)
	decision.ConstraintResult = "deny"

	envelope, err := BuildProofEnvelopeV10(execution, decision)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV10(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on invalid decision proof")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "decision proof invalid" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV10FailsOnExecutionDecisionBindingMismatch(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v10"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)
	decision.ActionID = "action.other.v1"

	envelope, err := BuildProofEnvelopeV10(execution, decision)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV10(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on execution/decision binding mismatch")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	expected := []string{
		"decision proof invalid",
		"execution_decision action_id mismatch",
	}

	if len(result.Issues) != len(expected) {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}

	for i := range expected {
		if result.Issues[i] != expected[i] {
			t.Fatalf("unexpected issues: %+v", result.Issues)
		}
	}
}

func TestVerifyChainV07Pass(t *testing.T) {
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

	r1, state1, _ := BuildTransitionV06(state0)
	r2, state2, _ := BuildTransitionV06(state1)

	chain := TransitionChainV07{
		States:   []CanonicalStateV06{state0, state1, state2},
		Receipts: []TransitionReceiptV06{r1, r2},
	}

	res, err := VerifyChainV07(chain)
	if err != nil {
		t.Fatal(err)
	}

	if !res.Match {
		t.Fatalf("expected PASS, got FAIL: %v", res.Issues)
	}
}
func TestVerifyChainV07FailsOnTamperedState(t *testing.T) {
	state0 := buildInitialState()

	r1, state1, _ := BuildTransitionV06(state0)
	r2, state2, _ := BuildTransitionV06(state1)

	// Manipulation
	state1.Refs["output.summary.v1"] = HashStringV06("tampered")

	chain := TransitionChainV07{
		States:   []CanonicalStateV06{state0, state1, state2},
		Receipts: []TransitionReceiptV06{r1, r2},
	}

	res, _ := VerifyChainV07(chain)

	if res.Match {
		t.Fatal("expected FAIL on tampered state")
	}
}

func TestVerifyChainV07FailsOnReorderedReceipts(t *testing.T) {
	state0 := buildInitialState()

	r1, state1, _ := BuildTransitionV06(state0)
	r2, state2, _ := BuildTransitionV06(state1)

	// Reihenfolge vertauscht
	chain := TransitionChainV07{
		States:   []CanonicalStateV06{state0, state1, state2},
		Receipts: []TransitionReceiptV06{r2, r1},
	}

	res, _ := VerifyChainV07(chain)

	if res.Match {
		t.Fatal("expected FAIL on reordered receipts")
	}
}

func TestVerifyChainV07FailsOnMissingReceipt(t *testing.T) {
	state0 := buildInitialState()

	_, state1, _ := BuildTransitionV06(state0)
	_, state2, _ := BuildTransitionV06(state1)

	chain := TransitionChainV07{
		States:   []CanonicalStateV06{state0, state1, state2},
		Receipts: []TransitionReceiptV06{}, // fehlt
	}

	res, _ := VerifyChainV07(chain)

	if res.Match {
		t.Fatal("expected FAIL on missing receipt")
	}
}

func TestVerifyChainV07FailsOnBrokenContinuity(t *testing.T) {
	state0 := buildInitialState()

	r1, state1, _ := BuildTransitionV06(state0)
	r2, state2, _ := BuildTransitionV06(state1)

	// Continuity brechen
	r2.PrevStateHash = "sha256:fake"

	chain := TransitionChainV07{
		States:   []CanonicalStateV06{state0, state1, state2},
		Receipts: []TransitionReceiptV06{r1, r2},
	}

	res, _ := VerifyChainV07(chain)

	if res.Match {
		t.Fatal("expected FAIL on broken continuity")
	}
}

func buildInitialState() CanonicalStateV06 {
	return CanonicalStateV06{
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
}
