package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
)

func runProject(cfg Config) error {
	return runProjectLike(cfg)
}

func runOperator(cfg Config) error {
	return runProjectLike(cfg)
}

func runProjectLike(cfg Config) error {
	api := NewHTTPClient()
	if err := hydrateProjectCredentials(&cfg); err != nil {
		return err
	}

	sessionToken, err := LoginWithPassword(api, cfg.InfisicalURL, cfg.InfisicalEmail, cfg.InfisicalPassword)
	if err != nil {
		return err
	}
	orgToken, err := SelectOrganization(api, cfg.InfisicalURL, sessionToken, cfg.OrganizationID)
	if err != nil {
		return err
	}
	headers := BearerHeaders(orgToken)

	project, err := EnsureProject(api, cfg.InfisicalURL, headers, cfg.ProjectName, cfg.ProjectSlug)
	if err != nil {
		return err
	}
	if err := EnsureEnvironment(api, cfg.InfisicalURL, headers, project, cfg.EnvironmentName, cfg.EnvironmentSlug); err != nil {
		return err
	}
	orgRole := ""
	if cfg.Mode == ModeOperator {
		orgRole = cfg.OrganizationIdentityRole
	}
	identityID, identityName, err := EnsureIdentity(api, cfg.InfisicalURL, headers, cfg.OrganizationID, cfg.IdentityName, orgRole)
	if err != nil {
		return err
	}
	if err := EnsureIdentityMembership(api, cfg.InfisicalURL, headers, project.ID, identityID, cfg.IdentityRole); err != nil {
		return err
	}

	var kube *HTTPClient
	var saToken string
	var caCert string
	if cfg.EnableKubernetesAuth || cfg.WriteKubernetesSecret || cfg.OutputStatusConfigMap != "" {
		var kubeErr error
		kube, saToken, caCert, kubeErr = NewKubeHTTPClient()
		if kubeErr != nil {
			return kubeErr
		}
	}

	if cfg.EnableKubernetesAuth {
		if err := EnsureKubernetesAuth(api, cfg.InfisicalURL, headers, identityID, cfg.KubernetesAuthHost, caCert, cfg.AllowedNamespaces, cfg.AllowedServiceAccounts); err != nil {
			return err
		}
	}

	universalAuthClientID := ""
	universalAuthClientSecret := ""
	if cfg.EnableUniversalAuth {
		universalAuthClientID, err = EnsureUniversalAuth(api, cfg.InfisicalURL, headers, identityID)
		if err != nil {
			return err
		}
		universalAuthClientSecret, err = CreateUniversalAuthClientSecret(api, cfg.InfisicalURL, headers, identityID)
		if err != nil {
			return err
		}
	}

	if cfg.SmokeTestSecretKey != "" && cfg.SmokeTestSecretValue != "" {
		if err := EnsureSecretValue(api, cfg.InfisicalURL, headers, project.ID, cfg.EnvironmentSlug, cfg.SmokeTestSecretKey, cfg.SmokeTestSecretValue, "/"); err != nil {
			return err
		}
	}

	for _, secret := range cfg.Secrets {
		if err := EnsureSecretValue(api, cfg.InfisicalURL, headers, project.ID, cfg.EnvironmentSlug, secret.Key, secret.Value, secret.Path); err != nil {
			return err
		}
	}

	if cfg.WriteKubernetesSecret {
		kubeAPI := fmt.Sprintf("https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), kubeServicePort())
		labels := map[string]string{
			"homelab.io/app":      cfg.ProjectSlug,
			"homelab.io/category": "platform",
		}

		if err := UpsertSecret(kube, kubeAPI, cfg.OutputSecretNamespace, saToken, cfg.OutputSecretName, cfg.OutputSecretKey, identityID, labels); err != nil {
			return err
		}

		if cfg.OutputProjectSecretName != "" && cfg.OutputProjectSecretKey != "" {
			if err := UpsertSecret(kube, kubeAPI, cfg.OutputSecretNamespace, saToken, cfg.OutputProjectSecretName, cfg.OutputProjectSecretKey, project.ID, labels); err != nil {
				return err
			}
		}
	}

	result := Result{
		ProjectID:                 project.ID,
		ProjectSlug:               project.Slug,
		IdentityID:                identityID,
		IdentityName:              identityName,
		EnvironmentSlug:           cfg.EnvironmentSlug,
		UniversalAuthClientID:     universalAuthClientID,
		UniversalAuthClientSecret: universalAuthClientSecret,
	}

	if err := writeProjectStatus(cfg, kube, saToken, result); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func hydrateProjectCredentials(cfg *Config) error {
	if cfg.InfisicalEmail != "" && cfg.InfisicalPassword != "" {
		return nil
	}

	kube, saToken, _, err := NewKubeHTTPClient()
	if err != nil {
		return err
	}
	kubeAPI := fmt.Sprintf("https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), kubeServicePort())
	data, err := GetSecretData(kube, kubeAPI, cfg.BootstrapSecretNamespace, saToken, cfg.BootstrapSecretName)
	if err != nil {
		return err
	}

	email := data[cfg.BootstrapEmailKey]
	password := data[cfg.BootstrapPasswordKey]
	if email == "" {
		return fmt.Errorf("missing %q in secret %s/%s", cfg.BootstrapEmailKey, cfg.BootstrapSecretNamespace, cfg.BootstrapSecretName)
	}
	if password == "" {
		return fmt.Errorf("missing %q in secret %s/%s", cfg.BootstrapPasswordKey, cfg.BootstrapSecretNamespace, cfg.BootstrapSecretName)
	}

	cfg.InfisicalEmail = email
	cfg.InfisicalPassword = password
	return nil
}

func writeProjectStatus(cfg Config, kube *HTTPClient, saToken string, result Result) error {
	if cfg.OutputStatusConfigMap == "" {
		return nil
	}
	if kube == nil {
		return fmt.Errorf("OUTPUT_STATUS_CONFIGMAP requires in-cluster Kubernetes access")
	}

	kubeAPI := fmt.Sprintf("https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), kubeServicePort())
	labels := map[string]string{
		"homelab.io/app":      cfg.ProjectSlug,
		"homelab.io/category": "platform",
	}

	return UpsertConfigMap(kube, kubeAPI, cfg.OutputSecretNamespace, saToken, cfg.OutputStatusConfigMap, map[string]string{
		"result":                   "bootstrapped",
		"message":                  projectStatusMessage(cfg.Mode),
		"mode":                     string(cfg.Mode),
		"projectId":                result.ProjectID,
		"projectSlug":              result.ProjectSlug,
		"identityId":               result.IdentityID,
		"identityName":             result.IdentityName,
		"environmentSlug":          result.EnvironmentSlug,
		"organizationIdentityRole": cfg.OrganizationIdentityRole,
		"projectIdentityRole":      cfg.IdentityRole,
	}, labels)
}

func projectStatusMessage(mode Mode) string {
	if mode == ModeOperator {
		return "Operator bootstrap completed."
	}
	return "Project bootstrap completed."
}
