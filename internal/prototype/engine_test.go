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

func buildV11Dependency() ExternalDependencyRefV11 {
	return ExternalDependencyRefV11{
		ID:          "api.external.v1",
		Type:        "api",
		Source:      "example.external.api",
		Fingerprint: HashStringV06("external-response-v11"),
		Boundary:    "declared-not-replayed",
	}
}

func TestVerifyProofEnvelopeV11Pass(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{buildV11Dependency()},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
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

func TestVerifyProofEnvelopeV11FailsOnMissingDependencies(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on missing external dependencies")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "external_dependencies missing" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV11FailsOnDependencyFingerprintMissing(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	dep := buildV11Dependency()
	dep.Fingerprint = ""

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{dep},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on missing dependency fingerprint")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "external_dependency fingerprint missing" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV11FailsOnDependencyTamper(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{buildV11Dependency()},
	)
	if err != nil {
		t.Fatal(err)
	}

	envelope.ExternalDependencies[0].Fingerprint = HashStringV06("tampered-external-response")

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on dependency tamper")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "envelope_hash mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV11FailsOnDependencyIDMissing(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	dep := buildV11Dependency()
	dep.ID = ""

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{dep},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on missing dependency id")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "external_dependency id missing" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV11FailsOnDependencyTypeInvalid(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	dep := buildV11Dependency()
	dep.Type = "random"

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{dep},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on invalid dependency type")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "external_dependency type invalid" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV11FailsOnDuplicateDependencyID(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	dep1 := buildV11Dependency()
	dep2 := buildV11Dependency()
	dep2.Fingerprint = HashStringV06("external-response-v11-second")

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{dep1, dep2},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on duplicate dependency id")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "external_dependency duplicate id" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyProofEnvelopeV11FailsOnOneInvalidDependencyAmongMany(t *testing.T) {
	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = "sha256:output-v11"

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	dep1 := buildV11Dependency()

	dep2 := ExternalDependencyRefV11{
		ID:          "human.approval.v1",
		Type:        "human",
		Source:      "reviewer.local",
		Fingerprint: "",
		Boundary:    "declared-not-replayed",
	}

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{dep1, dep2},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := VerifyProofEnvelopeV11(state0, envelope, state1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL when one dependency among many is invalid")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "external_dependency fingerprint missing" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func buildV12Envelope(t *testing.T, outputRef string, sequence string) (CanonicalStateV06, ProofEnvelopeV11, CanonicalStateV06) {
	t.Helper()

	state0 := buildInitialState()
	state1 := state0
	state1.Action.OutputRef = outputRef

	execution := buildV10ExecutionReceipt(t, state0, state1)
	decision := buildV10DecisionReceipt(t, state0, state1)

	decision.SequenceBoundary = sequence
	decision.PolicySetHash = HashStringV06("policy-set:v12")
	decision.AuthorizationContext = "authz-context:v12"
	decision.CapabilityScope = "summary:create"

	dep := ExternalDependencyRefV11{
		ID:          "api.external.v12",
		Type:        "api",
		Source:      "example.external.api",
		Fingerprint: HashStringV06("external-response-v12"),
		Boundary:    "declared-not-replayed",
	}

	envelope, err := BuildProofEnvelopeV11(
		execution,
		decision,
		[]ExternalDependencyRefV11{dep},
	)
	if err != nil {
		t.Fatal(err)
	}

	return state0, envelope, state1
}

func buildV12Link(t *testing.T, from ProofEnvelopeV11, to ProofEnvelopeV11, fromSeq int, toSeq int) CompositionLinkV12 {
	t.Helper()

	fromHash, err := HashProofEnvelopeV11(from)
	if err != nil {
		t.Fatal(err)
	}

	toHash, err := HashProofEnvelopeV11(to)
	if err != nil {
		t.Fatal(err)
	}

	return CompositionLinkV12{
		FromEnvelopeHash: fromHash,
		ToEnvelopeHash:   toHash,
		AuthorityContext: from.Decision.AuthorizationContext,
		DependencyScope:  dependencyScopeV12(from.ExternalDependencies),
		PolicySetHash:    from.Decision.PolicySetHash,
		SequenceFrom:     fromSeq,
		SequenceTo:       toSeq,
	}
}

func TestVerifyCompositionV12Pass(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %+v", result.Issues)
	}
}

func TestVerifyCompositionV12FailsOnAuthorityContinuityMismatch(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	envelope1.Decision.AuthorizationContext = "authz-context:changed"

	envelope1, err := BuildProofEnvelopeV11(
		envelope1.Execution,
		envelope1.Decision,
		envelope1.ExternalDependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on authority continuity mismatch")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "authority continuity mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV12FailsOnDependencyContinuityMismatch(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	envelope1.ExternalDependencies[0].Fingerprint = HashStringV06("changed-external-response")

	envelope1, err := BuildProofEnvelopeV11(
		envelope1.Execution,
		envelope1.Decision,
		envelope1.ExternalDependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on dependency continuity mismatch")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "dependency continuity mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV12FailsOnTemporalSequenceInvalid(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	link := buildV12Link(t, envelope0, envelope1, 2, 1)

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on temporal sequence invalid")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "temporal sequence invalid" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV12FailsOnLinkHashMismatch(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	link := buildV12Link(t, envelope0, envelope1, 1, 2)
	link.ToEnvelopeHash = "sha256:tampered"

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on link hash mismatch")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "composition to_envelope_hash mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV12FailsOnDuplicateEnvelope(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, _, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	link := buildV12Link(t, envelope0, envelope0, 1, 2)

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope0},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on duplicate envelope")
	}
}

func TestVerifyCompositionV12FailsOnSequenceGap(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:3")

	link := buildV12Link(t, envelope0, envelope1, 1, 3)

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on sequence gap")
	}
}

func TestVerifyCompositionV12FailsOnMissingLinkFields(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v12-b", "seq:2")

	link := buildV12Link(t, envelope0, envelope1, 1, 2)
	link.AuthorityContext = ""
	link.DependencyScope = ""
	link.PolicySetHash = ""

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on missing link fields")
	}
}

func TestVerifyCompositionV12PassesSingleEnvelopeWithoutLinks(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v12-a", "seq:1")

	chain := CompositionChainV12{
		Envelopes: []ProofEnvelopeV11{envelope0},
		Links:     []CompositionLinkV12{},
	}

	result, err := VerifyCompositionV12(
		[]CanonicalStateV06{prev0},
		chain,
		[]CanonicalStateV06{next0},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS for single envelope without links, got issues: %+v", result.Issues)
	}

	if result.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
}

func strictBoundaryV13() ContinuityBoundaryV13 {
	return ContinuityBoundaryV13{
		AuthorityInvariant:  true,
		PolicyInvariant:     true,
		CapabilityInvariant: true,
		DependencyInvariant: true,
		TemporalInvariant:   true,
	}
}

func relaxedBoundaryV13() ContinuityBoundaryV13 {
	return ContinuityBoundaryV13{
		AuthorityInvariant:  false,
		PolicyInvariant:     false,
		CapabilityInvariant: false,
		DependencyInvariant: false,
		TemporalInvariant:   false,
	}
}

func TestVerifyCompositionV13PassesWithStrictNoDrift(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:2")

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  strictBoundaryV13(),
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %+v", result.Issues)
	}
}

func TestVerifyCompositionV13FailsOnForbiddenAuthorityDrift(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:2")

	envelope1.Decision.AuthorizationContext = "authz-context:v13-changed"

	var err error
	envelope1, err = BuildProofEnvelopeV11(
		envelope1.Execution,
		envelope1.Decision,
		envelope1.ExternalDependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  strictBoundaryV13(),
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on forbidden authority drift")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "authority continuity mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV13PassesOnAllowedAuthorityDrift(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:2")

	envelope1.Decision.AuthorizationContext = "authz-context:v13-changed"

	var err error
	envelope1, err = BuildProofEnvelopeV11(
		envelope1.Execution,
		envelope1.Decision,
		envelope1.ExternalDependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	boundary := strictBoundaryV13()
	boundary.AuthorityInvariant = false

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  boundary,
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS on allowed authority drift, got issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV13FailsOnForbiddenDependencyDrift(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:2")

	envelope1.ExternalDependencies[0].Fingerprint = HashStringV06("changed-dependency-v13")

	var err error
	envelope1, err = BuildProofEnvelopeV11(
		envelope1.Execution,
		envelope1.Decision,
		envelope1.ExternalDependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  strictBoundaryV13(),
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on forbidden dependency drift")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "dependency continuity mismatch" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV13PassesOnAllowedDependencyDrift(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:2")

	envelope1.ExternalDependencies[0].Fingerprint = HashStringV06("changed-dependency-v13")

	var err error
	envelope1, err = BuildProofEnvelopeV11(
		envelope1.Execution,
		envelope1.Decision,
		envelope1.ExternalDependencies,
	)
	if err != nil {
		t.Fatal(err)
	}

	link := buildV12Link(t, envelope0, envelope1, 1, 2)

	boundary := strictBoundaryV13()
	boundary.DependencyInvariant = false

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  boundary,
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS on allowed dependency drift, got issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV13FailsOnForbiddenTemporalGap(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:3")

	link := buildV12Link(t, envelope0, envelope1, 1, 3)

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  strictBoundaryV13(),
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Match {
		t.Fatal("expected FAIL on forbidden temporal gap")
	}

	if len(result.Issues) != 1 || result.Issues[0] != "sequence continuity gap" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestVerifyCompositionV13PassesOnAllowedTemporalGap(t *testing.T) {
	prev0, envelope0, next0 := buildV12Envelope(t, "sha256:output-v13-a", "seq:1")
	prev1, envelope1, next1 := buildV12Envelope(t, "sha256:output-v13-b", "seq:3")

	link := buildV12Link(t, envelope0, envelope1, 1, 3)

	boundary := strictBoundaryV13()
	boundary.TemporalInvariant = false

	chain := CompositionChainV13{
		Envelopes: []ProofEnvelopeV11{envelope0, envelope1},
		Links:     []CompositionLinkV12{link},
		Boundary:  boundary,
	}

	result, err := VerifyCompositionV13(
		[]CanonicalStateV06{prev0, prev1},
		chain,
		[]CanonicalStateV06{next0, next1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Match {
		t.Fatalf("expected PASS on allowed temporal gap, got issues: %+v", result.Issues)
	}
}

func TestVerifyReceiptReferencesV14Pass(t *testing.T) {
	receipt := TransitionReceiptV08{
		InputRef:  "input.text.v1",
		PolicyRef: "policy.allow_summary.v1",
		OutputRef: "output.summary.v1",
	}

	refs := ReferenceSetV14{
		Inputs: map[string]bool{
			"input.text.v1": true,
		},
		Policies: map[string]bool{
			"policy.allow_summary.v1": true,
		},
		Outputs: map[string]bool{
			"output.summary.v1": true,
		},
	}

	result := VerifyReceiptReferencesV14(receipt, refs)

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %v", result.Issues)
	}
}

func TestVerifyReceiptReferencesV14FailsOnMissingRefs(t *testing.T) {
	receipt := TransitionReceiptV08{}

	refs := ReferenceSetV14{
		Inputs:   map[string]bool{},
		Policies: map[string]bool{},
		Outputs:  map[string]bool{},
	}

	result := VerifyReceiptReferencesV14(receipt, refs)

	if result.Match {
		t.Fatalf("expected FAIL")
	}

	if len(result.Issues) != 3 {
		t.Fatalf("expected 3 issues, got %d: %v", len(result.Issues), result.Issues)
	}
}

func TestVerifyReceiptReferencesV14FailsOnUnknownRefs(t *testing.T) {
	receipt := TransitionReceiptV08{
		InputRef:  "input.unknown",
		PolicyRef: "policy.unknown",
		OutputRef: "output.unknown",
	}

	refs := ReferenceSetV14{
		Inputs:   map[string]bool{},
		Policies: map[string]bool{},
		Outputs:  map[string]bool{},
	}

	result := VerifyReceiptReferencesV14(receipt, refs)

	if result.Match {
		t.Fatalf("expected FAIL")
	}

	if len(result.Issues) != 3 {
		t.Fatalf("expected 3 issues, got %d: %v", len(result.Issues), result.Issues)
	}
}

func TestValidatePolicyCompositionCase001PassWithOverride(t *testing.T) {
    receipts := []PolicyReceiptCase001{
        {StepID: "step_1", PolicyRef: "policy_v1", PolicyMode: "inherit", Status: "completed"},
        {StepID: "step_2", PolicyRef: "policy_v2", PolicyMode: "override", Status: "completed"},
    }

    result := ValidatePolicyCompositionCase001(receipts)

    if !result.Match {
        t.Fatalf("expected PASS, got FAIL: %v", result.Issues)
    }
}

func TestValidatePolicyCompositionCase001FailsOnUndeclaredPolicyDrift(t *testing.T) {
    receipts := []PolicyReceiptCase001{
        {StepID: "step_1", PolicyRef: "policy_v1", PolicyMode: "inherit", Status: "completed"},
        {StepID: "step_2", PolicyRef: "policy_v2", PolicyMode: "override", Status: "completed"},
        {StepID: "step_3", PolicyRef: "policy_v1", PolicyMode: "inherit", Status: "completed"},
    }

    result := ValidatePolicyCompositionCase001(receipts)

    if result.Match {
        t.Fatalf("expected FAIL")
    }

    if len(result.Issues) == 0 {
        t.Fatalf("expected policy drift issue")
    }
}
