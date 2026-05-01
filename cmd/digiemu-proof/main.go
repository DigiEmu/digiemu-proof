package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DigiEmu/digiemu-proof/internal/prototype"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	input := prototype.IntentEnvelope{
		Intent: "summarize_text",
		Context: map[string]string{
			"text": "DigiEmu Core verifies deterministic knowledge states.",
		},
	}

	switch os.Args[1] {
	case "run":
		snapshot := prototype.BuildSnapshot(input)
		hash, err := prototype.HashSnapshot(snapshot)
		exitOnErr(err)

		printJSON(map[string]any{
			"snapshot": snapshot,
			"hash":     hash,
		})

	case "verify":
		snapshot := prototype.BuildSnapshot(input)
		expectedHash, err := prototype.HashSnapshot(snapshot)
		exitOnErr(err)

		result, err := prototype.Verify(input, expectedHash)
		exitOnErr(err)

		printJSON(result)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	exitOnErr(err)
	fmt.Println(string(data))
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  digiemu-proof run")
	fmt.Println("  digiemu-proof verify")
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
