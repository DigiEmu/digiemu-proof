package prototype

type IntentEnvelope struct {
	Intent  string            `json:"intent"`
	Context map[string]string `json:"context"`
}

type PolicyResult struct {
	Policy string `json:"policy"`
	Result string `json:"result"`
}

type ActionResult struct {
	Type   string `json:"type"`
	Output string `json:"output"`
}

type ExecutionReceipt struct {
	Step   string `json:"step"`
	Result string `json:"result"`
}

type SnapshotV1 struct {
	Version string           `json:"version"`
	Input   IntentEnvelope   `json:"input"`
	Policy  PolicyResult     `json:"policy"`
	Action  ActionResult     `json:"action"`
	Receipt ExecutionReceipt `json:"receipt"`
}

type VerifyResult struct {
	Status       string `json:"status"`
	ExpectedHash string `json:"expected_hash"`
	ActualHash   string `json:"actual_hash"`
	Match        bool   `json:"match"`
}
