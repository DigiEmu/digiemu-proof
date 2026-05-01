package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DigiEmu/digiemu-proof/internal/prototype"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	inputPath := os.Args[2]

	input, err := readInput(inputPath)
	exitOnErr(err)

	switch command {
	case "run":
		snapshot := prototype.BuildSnapshot(input)
		hash, err := prototype.HashSnapshot(snapshot)
		exitOnErr(err)

		printJSON(map[string]any{
			"snapshot": snapshot,
			"hash":     hash,
		})

	case "verify":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "missing expected hash")
			printUsage()
			os.Exit(1)
		}

		expectedHash := os.Args[3]

		result, err := prototype.Verify(input, expectedHash)
		exitOnErr(err)

		printJSON(result)

	default:
		printUsage()
		os.Exit(1)
	}
}

func readInput(path string) (prototype.IntentEnvelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return prototype.IntentEnvelope{}, err
	}

	var input prototype.IntentEnvelope
	if err := json.Unmarshal(data, &input); err != nil {
		return prototype.IntentEnvelope{}, err
	}

	return input, nil
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	exitOnErr(err)
	fmt.Println(string(data))
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  digiemu-proof run input.json")
	fmt.Println("  digiemu-proof verify input.json sha256:<hash>")
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
