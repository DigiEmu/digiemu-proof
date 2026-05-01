package prototype

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func BuildSnapshot(input IntentEnvelope) SnapshotV1 {
	text := input.Context["text"]

	return SnapshotV1{
		Version: "snapshot.v1",
		Input:   input,
		Policy: PolicyResult{
			Policy: "allow_summary",
			Result: "pass",
		},
		Action: ActionResult{
			Type:   "summary",
			Output: text,
		},
		Receipt: ExecutionReceipt{
			Step:   "single_agent_summary",
			Result: "completed",
		},
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
