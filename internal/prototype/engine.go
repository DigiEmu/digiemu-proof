package prototype

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	SnapshotVersion = "snapshot.v1"

	PolicyAllowSummaryV1 = "policy.allow_summary.v1"
	PolicyRuleSummaryV1  = "intent == summarize_text && context.text != empty"

	StepSummaryV1 = "step.summary.v1"
	ActorLocalV1  = "agent.local.v1"
)

func BuildSnapshot(input IntentEnvelope) SnapshotV1 {
	policy := EvaluatePolicy(input)
	action := ExecuteAction(input, policy)
	receipt := BuildReceipt(policy, action)

	return SnapshotV1{
		Version: SnapshotVersion,
		Input:   input,
		Policy:  policy,
		Action:  action,
		Receipt: receipt,
	}
}

func EvaluatePolicy(input IntentEnvelope) PolicyResult {
	text := input.Context["text"]

	if input.Intent == "summarize_text" && text != "" {
		return PolicyResult{
			PolicyID:   PolicyAllowSummaryV1,
			Rule:       PolicyRuleSummaryV1,
			Decision:   "allow",
			ReasonCode: "TEXT_PRESENT",
		}
	}

	return PolicyResult{
		PolicyID:   PolicyAllowSummaryV1,
		Rule:       PolicyRuleSummaryV1,
		Decision:   "deny",
		ReasonCode: "TEXT_MISSING_OR_UNSUPPORTED_INTENT",
	}
}

func ExecuteAction(input IntentEnvelope, policy PolicyResult) ActionResult {
	if policy.Decision != "allow" {
		return ActionResult{
			Type:   "summary",
			Output: "",
		}
	}

	return ActionResult{
		Type:   "summary",
		Output: input.Context["text"],
	}
}

func BuildReceipt(policy PolicyResult, action ActionResult) ExecutionReceipt {
	status := "completed"
	if policy.Decision != "allow" {
		status = "blocked"
	}

	return ExecutionReceipt{
		StepID:     StepSummaryV1,
		Actor:      ActorLocalV1,
		ActionType: action.Type,
		InputRef:   "intent.context.text",
		PolicyRef:  policy.PolicyID,
		OutputRef:  "action.output",
		Status:     status,
	}
}

func CanonicalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func HashSnapshot(snapshot SnapshotV1) (string, error) {
	data, err := CanonicalJSON(snapshot)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func Verify(input IntentEnvelope, expectedHash string) (VerifyResult, error) {
	snapshot := BuildSnapshot(input)

	actualHash, err := HashSnapshot(snapshot)
	if err != nil {
		return VerifyResult{}, err
	}

	match := actualHash == expectedHash

	status := "FAIL"
	if match {
		status = "PASS"
	}

	return VerifyResult{
		Status:       status,
		ExpectedHash: expectedHash,
		ActualHash:   actualHash,
		Match:        match,
	}, nil
}

func HashCanonicalStateV06(state CanonicalStateV06) (string, error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(bytes)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}

func BuildTransitionV06(state CanonicalStateV06) (TransitionReceiptV06, CanonicalStateV06, error) {
	prevHash, err := HashCanonicalStateV06(state)
	if err != nil {
		return TransitionReceiptV06{}, CanonicalStateV06{}, err
	}

	nextState := DeriveNextStateV06(state)

	nextHash, err := HashCanonicalStateV06(nextState)
	if err != nil {
		return TransitionReceiptV06{}, CanonicalStateV06{}, err
	}

	receipt := TransitionReceiptV06{
		StepID:        "step.summary.v1",
		Actor:         "agent.local.v1",
		ActionType:    "summary",
		InputRef:      state.Intent.InputRef,
		PolicyRef:     state.Policy.ID,
		OutputRef:     state.Action.OutputRef,
		PrevStateHash: prevHash,
		NextStateHash: nextHash,
		Status:        "completed",
	}

	return receipt, nextState, nil
}

func VerifyTransitionV06(
	prevState CanonicalStateV06,
	receipt TransitionReceiptV06,
	nextState CanonicalStateV06,
) (TransitionVerifyResultV06, error) {
	issues := []string{}

	// --- ALL checks first ---

	prevHash, err := HashCanonicalStateV06(prevState)
	if err != nil {
		return TransitionVerifyResultV06{}, err
	}

	nextHash, err := HashCanonicalStateV06(nextState)
	if err != nil {
		return TransitionVerifyResultV06{}, err
	}

	if receipt.PrevStateHash != prevHash {
		issues = append(issues, "prev_state_hash mismatch")
	}

	if receipt.NextStateHash != nextHash {
		issues = append(issues, "next_state_hash mismatch")
	}

	if receipt.InputRef != prevState.Intent.InputRef {
		issues = append(issues, "input_ref mismatch")
	}

	if receipt.PolicyRef != prevState.Policy.ID {
		issues = append(issues, "policy_ref mismatch")
	}

	if receipt.OutputRef != prevState.Action.OutputRef {
		issues = append(issues, "output_ref mismatch")
	}

	// NEW: independent derivation check
	derivedNextState := DeriveNextStateV06(prevState)

	derivedNextHash, err := HashCanonicalStateV06(derivedNextState)
	if err != nil {
		return TransitionVerifyResultV06{}, err
	}

	if derivedNextHash != nextHash {
		issues = append(issues, "derived_next_state mismatch")
	}

	// existing check
	if prevState.Policy.Decision == "allow" && nextState.Refs[prevState.Action.OutputRef] == "" {
		issues = append(issues, "expected output ref missing in next state")
	}

	// --- FINAL evaluation AFTER all checks ---

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return TransitionVerifyResultV06{
		Status: status,
		Match:  match,
		Issues: issues,
	}, nil
}

func HashStringV06(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DeriveNextStateV06(state CanonicalStateV06) CanonicalStateV06 {
	nextState := state
	nextState.Refs = map[string]string{}

	if state.Refs != nil {
		for key, value := range state.Refs {
			nextState.Refs[key] = value
		}
	}

	inputHash := nextState.Refs["input.text.v1"]
	policyHash := nextState.Refs["policy.allow_summary.v1"]

	if inputHash != "" && policyHash != "" && nextState.Policy.Decision == "allow" {
		nextState.Refs["output.summary.v1"] = HashStringV06("DigiEmu Core verifies deterministic knowledge states.")
	}

	return nextState
}
