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

type TransitionChainV07 struct {
	States   []CanonicalStateV06    `json:"states"`
	Receipts []TransitionReceiptV06 `json:"receipts"`
}

type ChainVerifyResultV07 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}

type TransitionReceiptV08 struct {
	StepID     string `json:"step_id"`
	Actor      string `json:"actor"`
	ActionType string `json:"action_type"`

	IntentID       string `json:"intent_id"`
	PolicyID       string `json:"policy_id"`
	PolicyDecision string `json:"policy_decision"`
	ActionID       string `json:"action_id"`

	InputRef  string `json:"input_ref"`
	PolicyRef string `json:"policy_ref"`
	OutputRef string `json:"output_ref"`

	PrevStateHash string `json:"prev_state_hash"`
	NextStateHash string `json:"next_state_hash"`

	Status string `json:"status"`
}

type TransitionReceiptV09 struct {
	// --- Execution Surface ---
	StepID     string `json:"step_id"`
	Actor      string `json:"actor"`
	ActionType string `json:"action_type"`

	IntentID       string `json:"intent_id"`
	PolicyID       string `json:"policy_id"`
	PolicyDecision string `json:"policy_decision"`
	ActionID       string `json:"action_id"`

	InputRef  string `json:"input_ref"`
	PolicyRef string `json:"policy_ref"`
	OutputRef string `json:"output_ref"`

	PrevStateHash string `json:"prev_state_hash"`
	NextStateHash string `json:"next_state_hash"`

	// --- Decision Surface ---
	PolicySetHash         string `json:"policy_set_hash"`
	AuthorizationContext  string `json:"authorization_context"`
	ConstraintResult      string `json:"constraint_result"`
	ActorID               string `json:"actor_id"`
	CapabilityScope       string `json:"capability_scope"`
	SequenceBoundary      string `json:"sequence_boundary"`
	DependencyFingerprint string `json:"dependency_fingerprint"`

	Status string `json:"status"`
}

type ProofEnvelopeV10 struct {
	Execution    TransitionReceiptV08 `json:"execution"`
	Decision     TransitionReceiptV09 `json:"decision"`
	EnvelopeHash string               `json:"envelope_hash"`
}

type ProofEnvelopeVerifyResultV10 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}

type ExternalDependencyRefV11 struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // api, human, time, system, agent
	Source      string `json:"source"`
	Fingerprint string `json:"fingerprint"`
	Boundary    string `json:"boundary"`
}

type ProofEnvelopeV11 struct {
	Execution            TransitionReceiptV08       `json:"execution"`
	Decision             TransitionReceiptV09       `json:"decision"`
	ExternalDependencies []ExternalDependencyRefV11 `json:"external_dependencies"`
	EnvelopeHash         string                     `json:"envelope_hash"`
}

type ProofEnvelopeVerifyResultV11 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}

type CompositionLinkV12 struct {
	FromEnvelopeHash string `json:"from_envelope_hash"`
	ToEnvelopeHash   string `json:"to_envelope_hash"`

	AuthorityContext string `json:"authority_context"`
	DependencyScope  string `json:"dependency_scope"`
	PolicySetHash    string `json:"policy_set_hash"`

	SequenceFrom int `json:"sequence_from"`
	SequenceTo   int `json:"sequence_to"`
}

type CompositionChainV12 struct {
	Envelopes []ProofEnvelopeV11   `json:"envelopes"`
	Links     []CompositionLinkV12 `json:"links"`
}

type CompositionVerifyResultV12 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}

type ContinuityBoundaryV13 struct {
	AuthorityInvariant  bool `json:"authority_invariant"`
	PolicyInvariant     bool `json:"policy_invariant"`
	CapabilityInvariant bool `json:"capability_invariant"`
	DependencyInvariant bool `json:"dependency_invariant"`
	TemporalInvariant   bool `json:"temporal_invariant"`
}

type CompositionChainV13 struct {
	Envelopes []ProofEnvelopeV11    `json:"envelopes"`
	Links     []CompositionLinkV12  `json:"links"`
	Boundary  ContinuityBoundaryV13 `json:"boundary"`
}

type CompositionVerifyResultV13 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}

// --- v0.14 Reference Integrity ---

type ReferenceSetV14 struct {
	Inputs   map[string]bool `json:"inputs"`
	Policies map[string]bool `json:"policies"`
	Outputs  map[string]bool `json:"outputs"`
}

type ReferenceVerifyResultV14 struct {
	Status string   `json:"status"`
	Match  bool     `json:"match"`
	Issues []string `json:"issues"`
}

// --- Case 001 Policy Drift ---

type PolicyModeCase001 string

const (
	PolicyModeInheritCase001  PolicyModeCase001 = "inherit"
	PolicyModeOverrideCase001 PolicyModeCase001 = "override"
)

type PolicyReceiptCase001 struct {
	StepID     string `json:"step_id"`
	PolicyRef  string `json:"policy_ref"`
	PolicyMode string `json:"policy_mode"`
	Status     string `json:"status"`
}

// --- Case 002 Authority Drift ---

type AuthorityReceiptCase002 struct {
	StepID        string `json:"step_id"`
	PolicyRef     string `json:"policy_ref"`
	Authority     string `json:"authority"`
	AuthorityMode string `json:"authority_mode"`
}
