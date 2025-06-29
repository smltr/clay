package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read the file
	fmt.Println("reading file ", filename)
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Transpile to Racket
	fmt.Println("getting string")
	code := string(content)
	fmt.Println("tokenizing")
	tokens := tokenize(code)
	fmt.Println("parsing")
	tree := parse(tokens)
	fmt.Println("transpiling to racket")
	racketCode := transpile(tree)

	fmt.Printf("Transpiled code:\n%s\n\n", racketCode)

	// Run in Racket
	cmd := exec.Command("racket", "-e", racketCode)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Error running Racket: %v\n", err)
		fmt.Printf("Racket output: %s\n", string(output))
		os.Exit(1)
	}

	fmt.Printf("Output:\n%s", string(output))
}
