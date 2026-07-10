package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"infisical-bootstrap-job/internal/jobhandoff"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "agent vault launcher failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	metadataPath := jobhandoff.Env("BOOTSTRAP_METADATA_PATH", "/bootstrap/agent-vault.json")
	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	var metadata jobhandoff.Metadata
	for {
		value, err := jobhandoff.ReadMetadata(metadataPath)
		if err == nil {
			metadata = value
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for bootstrap metadata at %s", metadataPath)
		}
		time.Sleep(2 * time.Second)
	}

	if err := os.Setenv("INFISICAL_URL", metadata.InfisicalURL); err != nil {
		return err
	}
	if err := os.Setenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_ID", metadata.UniversalAuthClientID); err != nil {
		return err
	}
	if err := os.Setenv("INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET", metadata.UniversalAuthClientSecret); err != nil {
		return err
	}

	binary, err := exec.LookPath("agent-vault")
	if err != nil {
		return err
	}

	args := []string{"agent-vault", "server", "--host", "0.0.0.0"}
	return syscall.Exec(binary, args, os.Environ())
}
