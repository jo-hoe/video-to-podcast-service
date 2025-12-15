//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// Get the root directory (parent of .githooks)
	rootDir, err := filepath.Abs("..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting root directory: %v\n", err)
		os.Exit(1)
	}

	// Source and destination paths
	srcPath := filepath.Join(rootDir, ".githooks", "pre-commit")
	dstPath := filepath.Join(rootDir, ".git", "hooks", "pre-commit")

	// Read source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening source file: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating hooks directory: %v\n", err)
		os.Exit(1)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating destination file: %v\n", err)
		os.Exit(1)
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Pre-commit hook installed successfully")
}
