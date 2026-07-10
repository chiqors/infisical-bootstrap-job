package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"infisical-bootstrap-job/internal/jobhandoff"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "agent vault bootstrap failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	metadata, err := jobhandoff.ReadMetadata(jobhandoff.Env("BOOTSTRAP_METADATA_PATH", "/bootstrap/agent-vault.json"))
	if err != nil {
		return err
	}

	if err := waitForAgentVault(jobhandoff.Env("AGENT_VAULT_URL", "http://agent-vault:14321")); err != nil {
		return err
	}

	email := jobhandoff.MustEnv("AGENT_VAULT_OWNER_EMAIL")
	password := jobhandoff.MustEnv("AGENT_VAULT_OWNER_PASSWORD")
	vaultName := jobhandoff.MustEnv("AGENT_VAULT_SYNC_VAULT_NAME")
	address := jobhandoff.Env("AGENT_VAULT_URL", "http://agent-vault:14321")
	serviceName := jobhandoff.Env("AGENT_VAULT_PROXY_SERVICE_NAME", "echo-api")
	serviceHost := jobhandoff.Env("AGENT_VAULT_PROXY_SERVICE_HOST", "echo-api.local:8080")
	tokenKey := jobhandoff.Env("AGENT_VAULT_PROXY_TOKEN_KEY", "E2E_API_TOKEN")
	agentName := jobhandoff.Env("AGENT_VAULT_PROXY_AGENT_NAME", "e2e-proxy-agent")
	outputPath := jobhandoff.Env("AGENT_VAULT_PROXY_OUTPUT_PATH", "/data/agent-vault-proxy.json")
	passthroughServiceName := jobhandoff.Env("AGENT_VAULT_PASSTHROUGH_SERVICE_NAME", "echo-passthrough")
	passthroughServiceHost := jobhandoff.Env("AGENT_VAULT_PASSTHROUGH_SERVICE_HOST", "echo-passthrough.local:8080")
	passthroughTokenKey := jobhandoff.Env("AGENT_VAULT_PASSTHROUGH_TOKEN_KEY", "E2E_API_TOKEN")
	passthroughPlaceholder := jobhandoff.Env("AGENT_VAULT_PASSTHROUGH_PLACEHOLDER", "__e2e_api_token__")
	passthroughAgentName := jobhandoff.Env("AGENT_VAULT_PASSTHROUGH_AGENT_NAME", "e2e-passthrough-agent")
	passthroughOutputPath := jobhandoff.Env("AGENT_VAULT_PASSTHROUGH_OUTPUT_PATH", "/data/agent-vault-passthrough.json")

	if err := login(address, email, password); err != nil {
		if err := register(address, email, password); err != nil {
			return err
		}
		if err := login(address, email, password); err != nil {
			return err
		}
	}

	exists, err := vaultExists(vaultName)
	if err != nil {
		return err
	}
	if !exists {
		if err := runAgentVault(
			"vault", "create", vaultName,
			"--credential-store=infisical",
			"--infisical-project-id", metadata.ProjectID,
			"--infisical-environment", metadata.EnvironmentSlug,
			"--infisical-path", "/",
		); err != nil {
			return err
		}
	}

	if err := runAgentVault("vault", "credential-store", "sync", vaultName); err != nil {
		return err
	}

	if err := ensureProxyService(vaultName, serviceName, serviceHost, tokenKey); err != nil {
		return err
	}
	if err := ensurePassthroughService(vaultName, passthroughServiceName, passthroughServiceHost, passthroughTokenKey, passthroughPlaceholder); err != nil {
		return err
	}

	agentToken, err := ensureProxyAgentToken(agentName, vaultName)
	if err != nil {
		return err
	}
	passthroughAgentToken, err := ensureProxyAgentToken(passthroughAgentName, vaultName)
	if err != nil {
		return err
	}

	info := proxyAccessInfo{
		AgentName:        agentName,
		AgentToken:       agentToken,
		VaultName:        vaultName,
		ServiceName:      serviceName,
		ServiceHost:      serviceHost,
		TokenKey:         tokenKey,
		ProxyURL:         "http://localhost:14322",
		TargetURL:        "http://" + serviceHost + "/check",
		SampleProxyCurl:  fmt.Sprintf("curl -fsS -x http://localhost:14322 --proxy-user '%s:%s' http://%s/check", agentToken, vaultName, serviceHost),
		MetadataFilePath: outputPath,
	}
	if err := writeProxyAccessInfo(outputPath, info); err != nil {
		return err
	}
	passthroughInfo := passthroughAccessInfo{
		AgentName:           passthroughAgentName,
		AgentToken:          passthroughAgentToken,
		VaultName:           vaultName,
		ServiceName:         passthroughServiceName,
		ServiceHost:         passthroughServiceHost,
		TokenKey:            passthroughTokenKey,
		Placeholder:         passthroughPlaceholder,
		ProxyURL:            "http://localhost:14322",
		TargetURL:           "http://" + passthroughServiceHost + "/check",
		HeaderName:          "Authorization",
		HeaderValueTemplate: "Bearer " + passthroughPlaceholder,
		SampleProxyCurl:     fmt.Sprintf("curl -fsS -x http://localhost:14322 --proxy-user '%s:%s' -H 'Authorization: Bearer %s' http://%s/check", passthroughAgentToken, vaultName, passthroughPlaceholder, passthroughServiceHost),
		MetadataFilePath:    passthroughOutputPath,
	}
	if err := writePassthroughAccessInfo(passthroughOutputPath, passthroughInfo); err != nil {
		return err
	}

	fmt.Printf("Wrote proxy test info to %s\n", outputPath)
	fmt.Printf("Wrote passthrough test info to %s\n", passthroughOutputPath)
	return nil
}

func waitForAgentVault(address string) error {
	deadline := time.Now().Add(5 * time.Minute)
	for {
		cmd := exec.Command("wget", "-q", "--spider", address+"/health")
		if err := cmd.Run(); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %s/health", address)
		}
		time.Sleep(2 * time.Second)
	}
}

func register(address, email, password string) error {
	cmd := exec.Command("agent-vault", "auth", "register", "--address", address, "--email", email, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func login(address, email, password string) error {
	cmd := exec.Command("agent-vault", "auth", "login", "--address", address, "--email", email, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func vaultExists(name string) (bool, error) {
	cmd := exec.Command("agent-vault", "vault", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), name), nil
}

func runAgentVault(args ...string) error {
	cmd := exec.Command("agent-vault", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runAgentVaultOutput(args ...string) (string, error) {
	cmd := exec.Command("agent-vault", args...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

func ensureProxyService(vaultName, serviceName, serviceHost, tokenKey string) error {
	return runAgentVault(
		"vault", "service", "add",
		"--vault", vaultName,
		"--name", serviceName,
		"--host", serviceHost,
		"--auth-type", "bearer",
		"--token-key", tokenKey,
	)
}

func ensurePassthroughService(vaultName, serviceName, serviceHost, tokenKey, placeholder string) error {
	contents := fmt.Sprintf(`services:
  - name: %s
    host: %s
    auth:
      type: passthrough
    substitutions:
      - key: %s
        placeholder: %s
        in: [header]
`, serviceName, serviceHost, tokenKey, placeholder)

	file, err := os.CreateTemp("", "agent-vault-service-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(contents); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return runAgentVault("vault", "service", "add", "--vault", vaultName, "-f", file.Name())
}

func ensureProxyAgentToken(agentName, vaultName string) (string, error) {
	if _, err := runAgentVaultOutput("agent", "info", agentName); err == nil {
		return runAgentVaultOutput("agent", "rotate", agentName, "--token-only")
	}

	return runAgentVaultOutput("agent", "create", agentName, "--vault", vaultName+":proxy", "--token-only")
}

type proxyAccessInfo struct {
	AgentName        string `json:"agentName"`
	AgentToken       string `json:"agentToken"`
	VaultName        string `json:"vaultName"`
	ServiceName      string `json:"serviceName"`
	ServiceHost      string `json:"serviceHost"`
	TokenKey         string `json:"tokenKey"`
	ProxyURL         string `json:"proxyUrl"`
	TargetURL        string `json:"targetUrl"`
	SampleProxyCurl  string `json:"sampleProxyCurl"`
	MetadataFilePath string `json:"metadataFilePath"`
}

type passthroughAccessInfo struct {
	AgentName           string `json:"agentName"`
	AgentToken          string `json:"agentToken"`
	VaultName           string `json:"vaultName"`
	ServiceName         string `json:"serviceName"`
	ServiceHost         string `json:"serviceHost"`
	TokenKey            string `json:"tokenKey"`
	Placeholder         string `json:"placeholder"`
	ProxyURL            string `json:"proxyUrl"`
	TargetURL           string `json:"targetUrl"`
	HeaderName          string `json:"headerName"`
	HeaderValueTemplate string `json:"headerValueTemplate"`
	SampleProxyCurl     string `json:"sampleProxyCurl"`
	MetadataFilePath    string `json:"metadataFilePath"`
}

func writeProxyAccessInfo(path string, info proxyAccessInfo) error {
	return writeJSONFile(path, info)
}

func writePassthroughAccessInfo(path string, info passthroughAccessInfo) error {
	return writeJSONFile(path, info)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
