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
		StepID:        StepSummaryV1,
		Actor:         ActorLocalV1,
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

	derivedNextState := DeriveNextStateV06(prevState)

	derivedNextHash, err := HashCanonicalStateV06(derivedNextState)
	if err != nil {
		return TransitionVerifyResultV06{}, err
	}

	if derivedNextHash != nextHash {
		issues = append(issues, "derived_next_state mismatch")
	}

	if prevState.Policy.Decision == "allow" && nextState.Refs[prevState.Action.OutputRef] == "" {
		issues = append(issues, "expected output ref missing in next state")
	}

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

func VerifyChainV07(chain TransitionChainV07) (ChainVerifyResultV07, error) {
	issues := []string{}

	if len(chain.States) < 2 {
		return ChainVerifyResultV07{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"invalid chain length"},
		}, nil
	}

	if len(chain.Receipts) != len(chain.States)-1 {
		return ChainVerifyResultV07{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"invalid chain length"},
		}, nil
	}

	for i := 0; i < len(chain.Receipts); i++ {
		prev := chain.States[i]
		next := chain.States[i+1]
		r := chain.Receipts[i]

		res, err := VerifyTransitionV06(prev, r, next)
		if err != nil {
			return ChainVerifyResultV07{}, err
		}

		if !res.Match {
			issues = append(issues, "transition_"+r.StepID+" invalid")
		}

		prevHash, err := HashCanonicalStateV06(prev)
		if err != nil {
			return ChainVerifyResultV07{}, err
		}

		nextHash, err := HashCanonicalStateV06(next)
		if err != nil {
			return ChainVerifyResultV07{}, err
		}

		if r.PrevStateHash != prevHash {
			issues = append(issues, "prev_state continuity mismatch")
		}

		if r.NextStateHash != nextHash {
			issues = append(issues, "next_state continuity mismatch")
		}

		if i > 0 {
			prevReceipt := chain.Receipts[i-1]
			if prevReceipt.NextStateHash != r.PrevStateHash {
				issues = append(issues, "chain order broken")
			}
		}
	}

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return ChainVerifyResultV07{
		Status: status,
		Match:  match,
		Issues: issues,
	}, nil
}

func VerifyTransitionReceiptV08(
	prevState CanonicalStateV06,
	receipt TransitionReceiptV08,
	nextState CanonicalStateV06,
) (TransitionVerifyResultV06, error) {
	issues := []string{}

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

	if receipt.IntentID != prevState.Intent.ID {
		issues = append(issues, "intent_id mismatch")
	}

	if receipt.PolicyID != prevState.Policy.ID {
		issues = append(issues, "policy_id mismatch")
	}

	if receipt.PolicyDecision != prevState.Policy.Decision {
		issues = append(issues, "policy_decision mismatch")
	}

	if receipt.ActionID != prevState.Action.ID {
		issues = append(issues, "action_id mismatch")
	}

	if receipt.ActionType != prevState.Action.Type {
		issues = append(issues, "action_type mismatch")
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

	if len(issues) > 0 {
		return TransitionVerifyResultV06{
			Status: "FAIL",
			Match:  false,
			Issues: issues,
		}, nil
	}

	return TransitionVerifyResultV06{
		Status: "PASS",
		Match:  true,
		Issues: []string{},
	}, nil
}

func VerifyTransitionReceiptV09(
	prevState CanonicalStateV06,
	receipt TransitionReceiptV09,
	nextState CanonicalStateV06,
) (TransitionVerifyResultV06, error) {
	issues := []string{}

	prevHash, err := HashCanonicalStateV06(prevState)
	if err != nil {
		return TransitionVerifyResultV06{}, err
	}

	nextHash, err := HashCanonicalStateV06(nextState)
	if err != nil {
		return TransitionVerifyResultV06{}, err
	}

	// --- Execution Surface ---

	if receipt.PrevStateHash != prevHash {
		issues = append(issues, "prev_state_hash mismatch")
	}

	if receipt.NextStateHash != nextHash {
		issues = append(issues, "next_state_hash mismatch")
	}

	if receipt.IntentID != prevState.Intent.ID {
		issues = append(issues, "intent_id mismatch")
	}

	if receipt.PolicyID != prevState.Policy.ID {
		issues = append(issues, "policy_id mismatch")
	}

	if receipt.PolicyDecision != prevState.Policy.Decision {
		issues = append(issues, "policy_decision mismatch")
	}

	if receipt.ActionID != prevState.Action.ID {
		issues = append(issues, "action_id mismatch")
	}

	if receipt.ActionType != prevState.Action.Type {
		issues = append(issues, "action_type mismatch")
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

	// --- Decision Surface ---

	if receipt.PolicySetHash == "" {
		issues = append(issues, "policy_set_hash missing")
	}

	if receipt.AuthorizationContext == "" {
		issues = append(issues, "authorization_context missing")
	}

	if receipt.ConstraintResult == "" {
		issues = append(issues, "constraint_result missing")
	}

	if receipt.ActorID == "" {
		issues = append(issues, "actor_id missing")
	}

	if receipt.CapabilityScope == "" {
		issues = append(issues, "capability_scope missing")
	}

	if receipt.SequenceBoundary == "" {
		issues = append(issues, "sequence_boundary missing")
	}

	if receipt.DependencyFingerprint == "" {
		issues = append(issues, "dependency_fingerprint missing")
	}

	if receipt.ConstraintResult != "" && receipt.ConstraintResult != "allow" {
		issues = append(issues, "constraint_result not allow")
	}

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

func HashProofEnvelopeV10(envelope ProofEnvelopeV10) (string, error) {
	envelope.EnvelopeHash = ""

	bytes, err := json.Marshal(envelope)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(bytes)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}

func BuildProofEnvelopeV10(
	execution TransitionReceiptV08,
	decision TransitionReceiptV09,
) (ProofEnvelopeV10, error) {
	envelope := ProofEnvelopeV10{
		Execution: execution,
		Decision:  decision,
	}

	hash, err := HashProofEnvelopeV10(envelope)
	if err != nil {
		return ProofEnvelopeV10{}, err
	}

	envelope.EnvelopeHash = hash

	return envelope, nil
}

func VerifyProofEnvelopeV10(
	prevState CanonicalStateV06,
	envelope ProofEnvelopeV10,
	nextState CanonicalStateV06,
) (ProofEnvelopeVerifyResultV10, error) {
	issues := []string{}

	expectedHash, err := HashProofEnvelopeV10(envelope)
	if err != nil {
		return ProofEnvelopeVerifyResultV10{}, err
	}

	if envelope.EnvelopeHash != expectedHash {
		issues = append(issues, "envelope_hash mismatch")
	}

	execResult, err := VerifyTransitionReceiptV08(prevState, envelope.Execution, nextState)
	if err != nil {
		return ProofEnvelopeVerifyResultV10{}, err
	}

	if !execResult.Match {
		issues = append(issues, "execution proof invalid")
	}

	decisionResult, err := VerifyTransitionReceiptV09(prevState, envelope.Decision, nextState)
	if err != nil {
		return ProofEnvelopeVerifyResultV10{}, err
	}

	if !decisionResult.Match {
		issues = append(issues, "decision proof invalid")
	}

	if envelope.Execution.PrevStateHash != envelope.Decision.PrevStateHash {
		issues = append(issues, "execution_decision prev_state_hash mismatch")
	}

	if envelope.Execution.NextStateHash != envelope.Decision.NextStateHash {
		issues = append(issues, "execution_decision next_state_hash mismatch")
	}

	if envelope.Execution.IntentID != envelope.Decision.IntentID {
		issues = append(issues, "execution_decision intent_id mismatch")
	}

	if envelope.Execution.PolicyID != envelope.Decision.PolicyID {
		issues = append(issues, "execution_decision policy_id mismatch")
	}

	if envelope.Execution.PolicyDecision != envelope.Decision.PolicyDecision {
		issues = append(issues, "execution_decision policy_decision mismatch")
	}

	if envelope.Execution.ActionID != envelope.Decision.ActionID {
		issues = append(issues, "execution_decision action_id mismatch")
	}

	if envelope.Execution.ActionType != envelope.Decision.ActionType {
		issues = append(issues, "execution_decision action_type mismatch")
	}

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return ProofEnvelopeVerifyResultV10{
		Status: status,
		Match:  match,
		Issues: issues,
	}, nil
}

func HashProofEnvelopeV11(envelope ProofEnvelopeV11) (string, error) {
	envelope.EnvelopeHash = ""

	bytes, err := json.Marshal(envelope)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(bytes)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}

func BuildProofEnvelopeV11(
	execution TransitionReceiptV08,
	decision TransitionReceiptV09,
	dependencies []ExternalDependencyRefV11,
) (ProofEnvelopeV11, error) {
	envelope := ProofEnvelopeV11{
		Execution:            execution,
		Decision:             decision,
		ExternalDependencies: dependencies,
	}

	hash, err := HashProofEnvelopeV11(envelope)
	if err != nil {
		return ProofEnvelopeV11{}, err
	}

	envelope.EnvelopeHash = hash

	return envelope, nil
}

func VerifyProofEnvelopeV11(
	prevState CanonicalStateV06,
	envelope ProofEnvelopeV11,
	nextState CanonicalStateV06,
) (ProofEnvelopeVerifyResultV11, error) {

	issues := []string{}

	// --- Envelope Hash ---

	expectedHash, err := HashProofEnvelopeV11(envelope)
	if err != nil {
		return ProofEnvelopeVerifyResultV11{}, err
	}

	if envelope.EnvelopeHash != expectedHash {
		issues = append(issues, "envelope_hash mismatch")
	}

	// --- v0.10 Validation (Execution + Decision Binding) ---

	v10Envelope, err := BuildProofEnvelopeV10(envelope.Execution, envelope.Decision)
	if err != nil {
		return ProofEnvelopeVerifyResultV11{}, err
	}

	v10Envelope.EnvelopeHash, err = HashProofEnvelopeV10(v10Envelope)
	if err != nil {
		return ProofEnvelopeVerifyResultV11{}, err
	}

	v10Result, err := VerifyProofEnvelopeV10(prevState, v10Envelope, nextState)
	if err != nil {
		return ProofEnvelopeVerifyResultV11{}, err
	}

	if !v10Result.Match {
		issues = append(issues, "proof_envelope_v10 invalid")
	}

	// --- External Dependency Surface ---

	if len(envelope.ExternalDependencies) == 0 {
		issues = append(issues, "external_dependencies missing")
	}

	seenIDs := map[string]bool{}

	allowedTypes := map[string]bool{
		"api":    true,
		"human":  true,
		"time":   true,
		"system": true,
		"agent":  true,
	}

	for _, dep := range envelope.ExternalDependencies {

		if dep.ID == "" {
			issues = append(issues, "external_dependency id missing")
		} else {
			if seenIDs[dep.ID] {
				issues = append(issues, "external_dependency duplicate id")
			}
			seenIDs[dep.ID] = true
		}

		if dep.Type == "" {
			issues = append(issues, "external_dependency type missing")
		} else {
			if !allowedTypes[dep.Type] {
				issues = append(issues, "external_dependency type invalid")
			}
		}

		if dep.Source == "" {
			issues = append(issues, "external_dependency source missing")
		}

		if dep.Fingerprint == "" {
			issues = append(issues, "external_dependency fingerprint missing")
		}

		if dep.Boundary == "" {
			issues = append(issues, "external_dependency boundary missing")
		}
	}

	// --- Final Result ---

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return ProofEnvelopeVerifyResultV11{
		Status: status,
		Match:  match,
		Issues: issues,
	}, nil
}

func VerifyCompositionV12(
	prevStates []CanonicalStateV06,
	chain CompositionChainV12,
	nextStates []CanonicalStateV06,
) (CompositionVerifyResultV12, error) {
	issues := []string{}

	if len(chain.Envelopes) == 0 {
		return CompositionVerifyResultV12{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"composition envelopes missing"},
		}, nil
	}

	if len(prevStates) != len(chain.Envelopes) || len(nextStates) != len(chain.Envelopes) {
		return CompositionVerifyResultV12{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"composition state envelope length mismatch"},
		}, nil
	}

	if len(chain.Envelopes) == 1 && len(chain.Links) == 0 {
		result, err := VerifyProofEnvelopeV11(prevStates[0], chain.Envelopes[0], nextStates[0])
		if err != nil {
			return CompositionVerifyResultV12{}, err
		}

		return CompositionVerifyResultV12{
			Status: result.Status,
			Match:  result.Match,
			Issues: result.Issues,
		}, nil
	}

	if len(chain.Envelopes) > 1 && len(chain.Links) != len(chain.Envelopes)-1 {
		return CompositionVerifyResultV12{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"composition link length mismatch"},
		}, nil
	}

	envelopeHashes := make([]string, len(chain.Envelopes))
	seenEnvelopeHashes := map[string]bool{}

	for i, envelope := range chain.Envelopes {
		result, err := VerifyProofEnvelopeV11(prevStates[i], envelope, nextStates[i])
		if err != nil {
			return CompositionVerifyResultV12{}, err
		}

		if !result.Match {
			issues = append(issues, "envelope_"+envelope.EnvelopeHash+" invalid")
		}

		hash, err := HashProofEnvelopeV11(envelope)
		if err != nil {
			return CompositionVerifyResultV12{}, err
		}

		envelopeHashes[i] = hash

		if seenEnvelopeHashes[hash] {
			issues = append(issues, "duplicate envelope hash")
		}
		seenEnvelopeHashes[hash] = true

		if envelope.EnvelopeHash != hash {
			issues = append(issues, "envelope_hash mismatch")
		}
	}

	for i := 0; i < len(chain.Links); i++ {
		link := chain.Links[i]

		from := chain.Envelopes[i]
		to := chain.Envelopes[i+1]

		if link.FromEnvelopeHash != envelopeHashes[i] {
			issues = append(issues, "composition from_envelope_hash mismatch")
		}

		if link.ToEnvelopeHash != envelopeHashes[i+1] {
			issues = append(issues, "composition to_envelope_hash mismatch")
		}

		if link.AuthorityContext == "" {
			issues = append(issues, "link authority_context missing")
		}

		if link.DependencyScope == "" {
			issues = append(issues, "link dependency_scope missing")
		}

		if link.PolicySetHash == "" {
			issues = append(issues, "link policy_set_hash missing")
		}

		if from.Decision.AuthorizationContext != to.Decision.AuthorizationContext {
			issues = append(issues, "authority continuity mismatch")
		}

		if link.AuthorityContext != from.Decision.AuthorizationContext {
			issues = append(issues, "composition authority_context mismatch")
		}

		if from.Decision.PolicySetHash != to.Decision.PolicySetHash {
			issues = append(issues, "policy continuity mismatch")
		}

		if link.PolicySetHash != from.Decision.PolicySetHash {
			issues = append(issues, "composition policy_set_hash mismatch")
		}

		if from.Decision.CapabilityScope != to.Decision.CapabilityScope {
			issues = append(issues, "capability continuity mismatch")
		}

		if link.SequenceTo <= link.SequenceFrom {
			issues = append(issues, "temporal sequence invalid")
		} else if link.SequenceTo != link.SequenceFrom+1 {
			issues = append(issues, "sequence continuity gap")
		}

		if i > 0 {
			previousLink := chain.Links[i-1]
			if link.SequenceFrom != previousLink.SequenceTo {
				issues = append(issues, "temporal continuity mismatch")
			}
		}

		fromDependencyScope := dependencyScopeV12(from.ExternalDependencies)
		toDependencyScope := dependencyScopeV12(to.ExternalDependencies)

		if fromDependencyScope != toDependencyScope {
			issues = append(issues, "dependency continuity mismatch")
		}

		if link.DependencyScope != fromDependencyScope {
			issues = append(issues, "composition dependency_scope mismatch")
		}
	}

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return CompositionVerifyResultV12{
		Status: status,
		Match:  match,
		Issues: issues,
	}, nil
}

func dependencyScopeV12(dependencies []ExternalDependencyRefV11) string {
	if len(dependencies) == 0 {
		return ""
	}

	scope := ""

	for _, dep := range dependencies {
		scope += dep.ID + ":" + dep.Fingerprint + ";"
	}

	return HashStringV06(scope)
}

func VerifyCompositionV13(
	prevStates []CanonicalStateV06,
	chain CompositionChainV13,
	nextStates []CanonicalStateV06,
) (CompositionVerifyResultV13, error) {
	issues := []string{}

	if len(chain.Envelopes) == 0 {
		return CompositionVerifyResultV13{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"composition envelopes missing"},
		}, nil
	}

	if len(prevStates) != len(chain.Envelopes) || len(nextStates) != len(chain.Envelopes) {
		return CompositionVerifyResultV13{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"composition state envelope length mismatch"},
		}, nil
	}

	if len(chain.Envelopes) == 1 && len(chain.Links) == 0 {
		result, err := VerifyProofEnvelopeV11(prevStates[0], chain.Envelopes[0], nextStates[0])
		if err != nil {
			return CompositionVerifyResultV13{}, err
		}

		return CompositionVerifyResultV13{
			Status: result.Status,
			Match:  result.Match,
			Issues: result.Issues,
		}, nil
	}

	if len(chain.Envelopes) > 1 && len(chain.Links) != len(chain.Envelopes)-1 {
		return CompositionVerifyResultV13{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"composition link length mismatch"},
		}, nil
	}

	envelopeHashes := make([]string, len(chain.Envelopes))
	seenEnvelopeHashes := map[string]bool{}

	for i, envelope := range chain.Envelopes {
		result, err := VerifyProofEnvelopeV11(prevStates[i], envelope, nextStates[i])
		if err != nil {
			return CompositionVerifyResultV13{}, err
		}

		if !result.Match {
			issues = append(issues, "envelope_"+envelope.EnvelopeHash+" invalid")
		}

		hash, err := HashProofEnvelopeV11(envelope)
		if err != nil {
			return CompositionVerifyResultV13{}, err
		}

		envelopeHashes[i] = hash

		if seenEnvelopeHashes[hash] {
			issues = append(issues, "duplicate envelope hash")
		}
		seenEnvelopeHashes[hash] = true

		if envelope.EnvelopeHash != hash {
			issues = append(issues, "envelope_hash mismatch")
		}
	}

	for i := 0; i < len(chain.Links); i++ {
		link := chain.Links[i]

		from := chain.Envelopes[i]
		to := chain.Envelopes[i+1]

		if link.FromEnvelopeHash != envelopeHashes[i] {
			issues = append(issues, "composition from_envelope_hash mismatch")
		}

		if link.ToEnvelopeHash != envelopeHashes[i+1] {
			issues = append(issues, "composition to_envelope_hash mismatch")
		}

		if chain.Boundary.AuthorityInvariant {
			if link.AuthorityContext == "" {
				issues = append(issues, "link authority_context missing")
			}

			if from.Decision.AuthorizationContext != to.Decision.AuthorizationContext {
				issues = append(issues, "authority continuity mismatch")
			}

			if link.AuthorityContext != from.Decision.AuthorizationContext {
				issues = append(issues, "composition authority_context mismatch")
			}
		}

		if chain.Boundary.PolicyInvariant {
			if link.PolicySetHash == "" {
				issues = append(issues, "link policy_set_hash missing")
			}

			if from.Decision.PolicySetHash != to.Decision.PolicySetHash {
				issues = append(issues, "policy continuity mismatch")
			}

			if link.PolicySetHash != from.Decision.PolicySetHash {
				issues = append(issues, "composition policy_set_hash mismatch")
			}
		}

		if chain.Boundary.CapabilityInvariant {
			if from.Decision.CapabilityScope != to.Decision.CapabilityScope {
				issues = append(issues, "capability continuity mismatch")
			}
		}

		if chain.Boundary.DependencyInvariant {
			if link.DependencyScope == "" {
				issues = append(issues, "link dependency_scope missing")
			}

			fromDependencyScope := dependencyScopeV12(from.ExternalDependencies)
			toDependencyScope := dependencyScopeV12(to.ExternalDependencies)

			if fromDependencyScope != toDependencyScope {
				issues = append(issues, "dependency continuity mismatch")
			}

			if link.DependencyScope != fromDependencyScope {
				issues = append(issues, "composition dependency_scope mismatch")
			}
		}

		if link.SequenceTo <= link.SequenceFrom {
			issues = append(issues, "temporal sequence invalid")
		} else if chain.Boundary.TemporalInvariant && link.SequenceTo != link.SequenceFrom+1 {
			issues = append(issues, "sequence continuity gap")
		}

		if chain.Boundary.TemporalInvariant && i > 0 {
			previousLink := chain.Links[i-1]
			if link.SequenceFrom != previousLink.SequenceTo {
				issues = append(issues, "temporal continuity mismatch")
			}
		}
	}

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return CompositionVerifyResultV13{
		Status: status,
		Match:  match,
		Issues: issues,
	}, nil
}

func VerifyReceiptReferencesV14(
	receipt TransitionReceiptV08,
	refs ReferenceSetV14,
) ReferenceVerifyResultV14 {
	issues := []string{}

	if receipt.InputRef == "" {
		issues = append(issues, "input_ref missing")
	} else if !refs.Inputs[receipt.InputRef] {
		issues = append(issues, "input_ref unknown")
	}

	if receipt.PolicyRef == "" {
		issues = append(issues, "policy_ref missing")
	} else if !refs.Policies[receipt.PolicyRef] {
		issues = append(issues, "policy_ref unknown")
	}

	if receipt.OutputRef == "" {
		issues = append(issues, "output_ref missing")
	} else if !refs.Outputs[receipt.OutputRef] {
		issues = append(issues, "output_ref unknown")
	}

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return ReferenceVerifyResultV14{
		Status: status,
		Match:  match,
		Issues: issues,
	}
}

func ValidatePolicyCompositionCase001(receipts []PolicyReceiptCase001) CompositionVerifyResultV13 {
	issues := []string{}

	if len(receipts) == 0 {
		return CompositionVerifyResultV13{
			Status: "FAIL",
			Match:  false,
			Issues: []string{"receipts missing"},
		}
	}

	currentPolicy := receipts[0].PolicyRef

	if currentPolicy == "" {
		issues = append(issues, "initial policy_ref missing")
	}

	for i, receipt := range receipts {
		if receipt.StepID == "" {
			issues = append(issues, "step_id missing")
		}

		if receipt.PolicyRef == "" {
			issues = append(issues, "policy_ref missing")
		}

		if receipt.PolicyMode != "inherit" && receipt.PolicyMode != "override" {
			issues = append(issues, "policy_mode invalid")
		}

		if i == 0 {
			continue
		}

		if receipt.PolicyRef != currentPolicy {
			if receipt.PolicyMode != "override" {
				issues = append(
					issues,
					"policy drift at "+receipt.StepID+": expected inherit from "+currentPolicy+", got "+receipt.PolicyRef,
				)
			} else {
				currentPolicy = receipt.PolicyRef
			}
		}
	}

	match := len(issues) == 0
	status := "FAIL"
	if match {
		status = "PASS"
	}

	return CompositionVerifyResultV13{
		Status: status,
		Match:  match,
		Issues: issues,
	}
}
