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
	kube, saToken, kubeAPI, err := maybeKubeWriters(cfg)
	if err != nil {
		return err
	}

	resp, err := BootstrapInstance(api, cfg.InfisicalURL, BootstrapInstanceRequest{
		Email:        cfg.BootstrapEmail,
		Password:     cfg.BootstrapPassword,
		Organization: cfg.OrganizationName,
	})
	if err != nil {
		if cfg.IgnoreIfBootstrapped && IsAlreadyBootstrappedError(err) {
			if err := writePlatformStatus(cfg, kube, kubeAPI, saToken, map[string]string{
				"result":               "already-set-up",
				"message":              "Platform bootstrap skipped because the Infisical instance was already set up.",
				"mode":                 string(cfg.Mode),
				"infisicalUrl":         cfg.InfisicalURL,
				"tokenSecretRequested": boolString(cfg.WriteKubernetesSecret),
				"tokenSecretWritten":   "false",
				"tokenSecretName":      cfg.OutputSecretName,
				"tokenSecretNamespace": cfg.OutputSecretNamespace,
				"tokenSecretReason":    "instance-already-set-up",
			}); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if cfg.WriteKubernetesSecret {
		if err := writeBootstrapOutputs(cfg, kube, kubeAPI, saToken, resp); err != nil {
			return err
		}
	}

	tokenSecretWritten := "false"
	tokenSecretReason := "not-requested"
	if cfg.WriteKubernetesSecret {
		tokenSecretWritten = "true"
		tokenSecretReason = "written"
	}

	if err := writePlatformStatus(cfg, kube, kubeAPI, saToken, map[string]string{
		"result":               "bootstrapped",
		"message":              "Platform bootstrap completed and returned a bootstrap identity token.",
		"mode":                 string(cfg.Mode),
		"infisicalUrl":         cfg.InfisicalURL,
		"organizationId":       resp.Organization.ID,
		"organizationName":     resp.Organization.Name,
		"organizationSlug":     resp.Organization.Slug,
		"identityId":           resp.Identity.ID,
		"identityName":         resp.Identity.Name,
		"userEmail":            resp.User.Email,
		"tokenSecretRequested": boolString(cfg.WriteKubernetesSecret),
		"tokenSecretWritten":   tokenSecretWritten,
		"tokenSecretName":      cfg.OutputSecretName,
		"tokenSecretNamespace": cfg.OutputSecretNamespace,
		"tokenSecretReason":    tokenSecretReason,
	}); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(resp)
}

func IsAlreadyBootstrappedError(err error) bool {
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

func writeBootstrapOutputs(cfg Config, kube *HTTPClient, kubeAPI, saToken string, resp BootstrapInstanceResponse) error {
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

func writePlatformStatus(cfg Config, kube *HTTPClient, kubeAPI, saToken string, data map[string]string) error {
	if cfg.OutputStatusConfigMap == "" {
		return nil
	}
	if kube == nil {
		return fmt.Errorf("OUTPUT_STATUS_CONFIGMAP requires in-cluster Kubernetes access")
	}

	labels := map[string]string{
		"homelab.io/app":      "infisical",
		"homelab.io/category": "platform",
	}
	return UpsertConfigMap(kube, kubeAPI, cfg.OutputSecretNamespace, saToken, cfg.OutputStatusConfigMap, data, labels)
}

func maybeKubeWriters(cfg Config) (*HTTPClient, string, string, error) {
	if !cfg.WriteKubernetesSecret && cfg.OutputStatusConfigMap == "" {
		return nil, "", "", nil
	}

	kube, saToken, _, err := NewKubeHTTPClient()
	if err != nil {
		return nil, "", "", err
	}
	kubeAPI := fmt.Sprintf("https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), kubeServicePort())
	return kube, saToken, kubeAPI, nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
