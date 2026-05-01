package prototype

func ValidateReceiptReferences(snapshot SnapshotV1) []ValidationIssue {
	issues := []ValidationIssue{}

	for _, receipt := range snapshot.Receipts {
		if !refExists(snapshot, receipt.InputRef) {
			issues = append(issues, issue(receipt.StepID, "input_ref", receipt.InputRef))
		}

		if !refExists(snapshot, receipt.PolicyRef) {
			issues = append(issues, issue(receipt.StepID, "policy_ref", receipt.PolicyRef))
		}

		if !refExists(snapshot, receipt.OutputRef) {
			issues = append(issues, issue(receipt.StepID, "output_ref", receipt.OutputRef))
		}
	}

	return issues
}

func refExists(snapshot SnapshotV1, ref string) bool {
	switch ref {
	case "intent.context.text":
		_, exists := snapshot.Input.Context["text"]
		return exists

	case "policy.allow_summary.v1":
		return snapshot.Policy.PolicyID == "policy.allow_summary.v1"

	case "policy.decision":
		return snapshot.Policy.Decision != ""

	case "action.output":
		return true

	case "validation.result":
		return true

	default:
		return false
	}
}

func issue(stepID, field, ref string) ValidationIssue {
	return ValidationIssue{
		ReceiptStepID: stepID,
		Field:         field,
		Ref:           ref,
		Reason:        "reference_not_found",
	}
}
