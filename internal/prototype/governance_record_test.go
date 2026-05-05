package prototype

import "testing"

func buildGovernanceRecordCase005() GovernanceRecordCase005 {
	return GovernanceRecordCase005{
		RecordID: "governance-record-case-005",
		Steps: []GovernanceRecordStepCase005{
			{
				StepID:                 "step_1",
				Stage:                  "check",
				DeclaredPolicyRef:      "policy_v1",
				DeclaredAuthority:      "reviewer_A",
				DeclaredContinuityMode: "inherit",
				PolicyFingerprint:      "policy_fp_v1",
				DependencyFingerprint:  "dependency_fp_v1",
				AuthorityAnchor:        "authority_anchor_A",
			},
			{
				StepID:                 "step_2",
				Stage:                  "review",
				DeclaredPolicyRef:      "policy_v1",
				DeclaredAuthority:      "reviewer_A",
				DeclaredContinuityMode: "inherit",
				PolicyFingerprint:      "policy_fp_v1",
				DependencyFingerprint:  "dependency_fp_v1",
				AuthorityAnchor:        "authority_anchor_A",
			},
			{
				StepID:                 "step_3",
				Stage:                  "approve",
				DeclaredPolicyRef:      "policy_v1",
				DeclaredAuthority:      "reviewer_A",
				DeclaredContinuityMode: "inherit",
				PolicyFingerprint:      "policy_fp_v1",
				DependencyFingerprint:  "dependency_fp_v1",
				AuthorityAnchor:        "authority_anchor_A",
			},
		},
	}
}

func TestValidateGovernanceRecordContinuityCase005PassesWithInheritedContinuity(t *testing.T) {
	record := buildGovernanceRecordCase005()

	result := ValidateGovernanceRecordContinuityCase005(record)

	if !result.Match {
		t.Fatalf("expected PASS, got FAIL: %v", result.Issues)
	}

	if result.Status != "PASS" {
		t.Fatalf("expected PASS, got %s", result.Status)
	}
}

func TestValidateGovernanceRecordContinuityCase005FailsOnDependencyMutationWithInherit(t *testing.T) {
	record := buildGovernanceRecordCase005()
	record.Steps[2].DependencyFingerprint = "dependency_fp_MUTATED"

	result := ValidateGovernanceRecordContinuityCase005(record)

	if result.Match {
		t.Fatal("expected FAIL on dependency mutation with inherit")
	}

	if result.Status != "FAIL" {
		t.Fatalf("expected FAIL, got %s", result.Status)
	}

	if len(result.Issues) != 1 || result.Issues[0] != "dependency_fingerprint drift on inherit" {
		t.Fatalf("unexpected issues: %v", result.Issues)
	}
}

func TestValidateGovernanceRecordContinuityCase005PassesOnDependencyMutationWithOverride(t *testing.T) {
	record := buildGovernanceRecordCase005()
	record.Steps[2].DeclaredContinuityMode = "override"
	record.Steps[2].DependencyFingerprint = "dependency_fp_v2"

	result := ValidateGovernanceRecordContinuityCase005(record)

	if !result.Match {
		t.Fatalf("expected PASS on declared override, got FAIL: %v", result.Issues)
	}
}

func TestValidateGovernanceRecordContinuityCase005FailsOnAuthorityDriftWithInherit(t *testing.T) {
	record := buildGovernanceRecordCase005()
	record.Steps[2].DeclaredAuthority = "reviewer_B"

	result := ValidateGovernanceRecordContinuityCase005(record)

	if result.Match {
		t.Fatal("expected FAIL on authority drift with inherit")
	}
}

func TestValidateGovernanceRecordContinuityCase005FailsOnPolicyDriftWithInherit(t *testing.T) {
	record := buildGovernanceRecordCase005()
	record.Steps[2].DeclaredPolicyRef = "policy_v2"

	result := ValidateGovernanceRecordContinuityCase005(record)

	if result.Match {
		t.Fatal("expected FAIL on policy drift with inherit")
	}
}
