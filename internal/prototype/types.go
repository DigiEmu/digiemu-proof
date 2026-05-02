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

// --- v0.6 Canonical State ---

type CanonicalStateV06 struct {
	SchemaVersion string            `json:"schema_version"`
	Intent        IntentRef         `json:"intent"`
	Policy        PolicyRef         `json:"policy"`
	Action        ActionRef         `json:"action"`
	Refs          map[string]string `json:"refs"`
}

type IntentRef struct {
	ID       string `json:"id"`
	InputRef string `json:"input_ref"`
}

type PolicyRef struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

type ActionRef struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	OutputRef string `json:"output_ref"`
}

type TransitionReceiptV06 struct {
	StepID        string `json:"step_id"`
	Actor         string `json:"actor"`
	ActionType    string `json:"action_type"`
	InputRef      string `json:"input_ref"`
	PolicyRef     string `json:"policy_ref"`
	OutputRef     string `json:"output_ref"`
	PrevStateHash string `json:"prev_state_hash"`
	NextStateHash string `json:"next_state_hash"`
	Status        string `json:"status"`
}

type TransitionVerifyResultV06 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}
