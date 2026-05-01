package prototype

type IntentEnvelope struct {
	Intent  string            `json:"intent"`
	Context map[string]string `json:"context"`
}

type PolicyResult struct {
	PolicyID   string `json:"policy_id"`
	Rule       string `json:"rule"`
	Decision   string `json:"decision"`
	ReasonCode string `json:"reason_code"`
}

type ActionResult struct {
	Type   string `json:"type"`
	Output string `json:"output"`
}

type ExecutionReceipt struct {
	StepID     string `json:"step_id"`
	Actor      string `json:"actor"`
	ActionType string `json:"action_type"`
	InputRef   string `json:"input_ref"`
	PolicyRef  string `json:"policy_ref"`
	OutputRef  string `json:"output_ref"`
	Status     string `json:"status"`
}

type SnapshotV1 struct {
	Version  string             `json:"version"`
	Input    IntentEnvelope     `json:"input"`
	Policy   PolicyResult       `json:"policy"`
	Action   ActionResult       `json:"action"`
	Receipts []ExecutionReceipt `json:"receipts"`
}

type VerifyResult struct {
	Status       string `json:"status"`
	ExpectedHash string `json:"expected_hash"`
	ActualHash   string `json:"actual_hash"`
	Match        bool   `json:"match"`
}

type AuditMetadata struct {
	CreatedBy   string `json:"created_by"`
	Reviewer    string `json:"reviewer"`
	Note        string `json:"note"`
	Environment string `json:"environment"`
}

type ProofPackage struct {
	Snapshot SnapshotV1    `json:"snapshot"`
	Hash     string        `json:"hash"`
	Metadata AuditMetadata `json:"metadata"`
}

type ValidationIssue struct {
	ReceiptStepID string `json:"receipt_step_id"`
	Field         string `json:"field"`
	Ref           string `json:"ref"`
	Reason        string `json:"reason"`
}
