package bootstrap

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	InfisicalURL            string
	InfisicalEmail          string
	InfisicalPassword       string
	OrganizationID          string
	ProjectName             string
	ProjectSlug             string
	EnvironmentName         string
	EnvironmentSlug         string
	IdentityName            string
	IdentityRole            string
	EnableKubernetesAuth    bool
	KubernetesAuthHost      string
	AllowedNamespaces       string
	AllowedServiceAccounts  string
	WriteKubernetesSecret   bool
	OutputSecretNamespace   string
	OutputSecretName        string
	OutputSecretKey         string
	OutputProjectSecretName string
	OutputProjectSecretKey  string
	SmokeTestSecretKey      string
	SmokeTestSecretValue    string
}

func LoadConfig() Config {
	cfg := Config{
		InfisicalURL:            strings.TrimRight(mustEnv("INFISICAL_URL"), "/"),
		InfisicalEmail:          mustEnv("INFISICAL_EMAIL"),
		InfisicalPassword:       mustEnv("INFISICAL_PASSWORD"),
		OrganizationID:          mustEnv("ORGANIZATION_ID"),
		ProjectName:             mustEnv("PROJECT_NAME"),
		ProjectSlug:             mustEnv("PROJECT_SLUG"),
		EnvironmentName:         mustEnv("ENVIRONMENT_NAME"),
		EnvironmentSlug:         mustEnv("ENVIRONMENT_SLUG"),
		IdentityName:            mustEnv("IDENTITY_NAME"),
		IdentityRole:            mustEnv("IDENTITY_ROLE"),
		EnableKubernetesAuth:    envBool("ENABLE_KUBERNETES_AUTH", false),
		WriteKubernetesSecret:   envBool("WRITE_KUBERNETES_SECRET", false),
		OutputProjectSecretName: strings.TrimSpace(os.Getenv("OUTPUT_PROJECT_SECRET_NAME")),
		OutputProjectSecretKey:  strings.TrimSpace(os.Getenv("OUTPUT_PROJECT_SECRET_KEY")),
		SmokeTestSecretKey:      strings.TrimSpace(os.Getenv("SMOKE_TEST_SECRET_KEY")),
		SmokeTestSecretValue:    strings.TrimSpace(os.Getenv("SMOKE_TEST_SECRET_VALUE")),
	}

	if cfg.EnableKubernetesAuth {
		cfg.KubernetesAuthHost = mustEnv("KUBERNETES_AUTH_HOST")
		cfg.AllowedNamespaces = mustEnv("ALLOWED_NAMESPACES")
		cfg.AllowedServiceAccounts = mustEnv("ALLOWED_SERVICE_ACCOUNTS")
	}

	if cfg.WriteKubernetesSecret {
		cfg.OutputSecretNamespace = mustEnv("OUTPUT_SECRET_NAMESPACE")
		cfg.OutputSecretName = mustEnv("OUTPUT_SECRET_NAME")
		cfg.OutputSecretKey = mustEnv("OUTPUT_SECRET_KEY")
	}

	return cfg
}

func mustEnv(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		panic(fmt.Sprintf("missing required environment variable: %s", name))
	}
	return value
}

func envBool(name string, fallback bool) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
