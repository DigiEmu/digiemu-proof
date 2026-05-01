package prototype

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func BuildSnapshot(input IntentEnvelope) SnapshotV1 {
	text := input.Context["text"]

	policy := evaluatePolicy(input)

	action := ActionResult{
		Type:   "summary",
		Output: text,
	}

	receipt := ExecutionReceipt{
		StepID:     "step.summary.v1",
		Actor:      "agent.local.v1",
		ActionType: "summary",
		InputRef:   "intent.context.text",
		PolicyRef:  policy.PolicyID,
		OutputRef:  "action.output",
		Status:     "completed",
	}

	return SnapshotV1{
		Version: "snapshot.v1",
		Input:   input,
		Policy:  policy,
		Action:  action,
		Receipt: receipt,
	}
}

func evaluatePolicy(input IntentEnvelope) PolicyResult {
	text, exists := input.Context["text"]

	if input.Intent == "summarize_text" && exists && text != "" {
		return PolicyResult{
			PolicyID:   "policy.allow_summary.v1",
			Rule:       "intent == summarize_text && context.text != empty",
			Decision:   "allow",
			ReasonCode: "TEXT_PRESENT",
		}
	}

	return PolicyResult{
		PolicyID:   "policy.allow_summary.v1",
		Rule:       "intent == summarize_text && context.text != empty",
		Decision:   "deny",
		ReasonCode: "TEXT_MISSING",
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

func BuildProofPackage(input IntentEnvelope, metadata AuditMetadata) (ProofPackage, error) {
	snapshot := BuildSnapshot(input)

	hash, err := HashSnapshot(snapshot)
	if err != nil {
		return ProofPackage{}, err
	}

	return ProofPackage{
		Snapshot: snapshot,
		Hash:     hash,
		Metadata: metadata,
	}, nil
}
