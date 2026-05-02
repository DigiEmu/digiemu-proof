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
