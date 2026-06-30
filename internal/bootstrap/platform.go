package bootstrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func runPlatform(cfg Config) error {
	api := NewHTTPClient()

	resp, err := BootstrapInstance(api, cfg.InfisicalURL, BootstrapInstanceRequest{
		Email:        cfg.BootstrapEmail,
		Password:     cfg.BootstrapPassword,
		Organization: cfg.OrganizationName,
	})
	if err != nil {
		if cfg.IgnoreIfBootstrapped && isAlreadyBootstrappedError(err) {
			return nil
		}
		return err
	}

	if cfg.WriteKubernetesSecret {
		if err := writeBootstrapOutputs(cfg, resp); err != nil {
			return err
		}
	}

	return json.NewEncoder(os.Stdout).Encode(resp)
}

func isAlreadyBootstrappedError(err error) bool {
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}

	body := strings.ToLower(httpErr.Body)
	if httpErr.StatusCode == 400 || httpErr.StatusCode == 403 || httpErr.StatusCode == 422 {
		return strings.Contains(body, "already bootstrapped") ||
			strings.Contains(body, "has already been bootstrapped") ||
			strings.Contains(body, "instance already bootstrapped") ||
			strings.Contains(body, "admin account has already been created") ||
			strings.Contains(body, "instance has already been set up")
	}

	return false
}

func writeBootstrapOutputs(cfg Config, resp BootstrapInstanceResponse) error {
	kube, saToken, _, err := NewKubeHTTPClient()
	if err != nil {
		return err
	}

	kubeAPI := fmt.Sprintf("https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), kubeServicePort())
	labels := map[string]string{
		"homelab.io/app":      "infisical",
		"homelab.io/category": "platform",
	}

	return UpsertSecret(
		kube,
		kubeAPI,
		cfg.OutputSecretNamespace,
		saToken,
		cfg.OutputSecretName,
		cfg.OutputSecretKey,
		resp.Identity.Credentials.Token,
		labels,
	)
}
