package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestTBNBoundarySnapshotHash(t *testing.T) {
	raw, err := os.ReadFile("testdata/tbn_boundary_snapshot.json")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	var snapshot any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("failed to parse snapshot JSON: %v", err)
	}

	// Go encoding/json serializes map keys in deterministic sorted order.
	canonical, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("failed to canonicalize snapshot: %v", err)
	}

	sum := sha256.Sum256(canonical)
	got := "sha256:" + hex.EncodeToString(sum[:])

	want := "sha256:cdddcf74e59559fa53d482ebd4cd2ac645aba3606d307a650b44997c0560bcdc"

	if got != want {
		t.Fatalf("hash mismatch\nwant: %s\ngot:  %s\ncanonical JSON:\n%s", want, got, string(canonical))
	}

	t.Logf("DigiEmu Proof hash confirmed: %s", got)
}
