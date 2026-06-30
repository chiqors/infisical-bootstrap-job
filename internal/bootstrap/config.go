package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Mode string

const (
	ModeProject  Mode = "project"
	ModeOperator Mode = "operator"
	ModePlatform Mode = "platform"
)

type Config struct {
	Mode                     Mode
	InfisicalURL             string
	BootstrapEmail           string
	BootstrapPassword        string
	OrganizationName         string
	IgnoreIfBootstrapped     bool
	InfisicalEmail           string
	InfisicalPassword        string
	BootstrapSecretNamespace string
	BootstrapSecretName      string
	BootstrapEmailKey        string
	BootstrapPasswordKey     string
	OrganizationID           string
	ProjectName              string
	ProjectSlug              string
	EnvironmentName          string
	EnvironmentSlug          string
	IdentityName             string
	OrganizationIdentityRole string
	IdentityRole             string
	EnableKubernetesAuth     bool
	KubernetesAuthHost       string
	AllowedNamespaces        string
	AllowedServiceAccounts   string
	WriteKubernetesSecret    bool
	OutputSecretNamespace    string
	OutputSecretName         string
	OutputSecretKey          string
	OutputStatusConfigMap    string
	OutputProjectSecretName  string
	OutputProjectSecretKey   string
	SmokeTestSecretKey       string
	SmokeTestSecretValue     string
	Secrets                  []SecretSpec
}

func LoadConfig() Config {
	cfg := Config{
		Mode:                     loadMode(),
		InfisicalURL:             strings.TrimRight(mustEnv("INFISICAL_URL"), "/"),
		BootstrapEmail:           strings.TrimSpace(os.Getenv("BOOTSTRAP_EMAIL")),
		BootstrapPassword:        strings.TrimSpace(os.Getenv("BOOTSTRAP_PASSWORD")),
		OrganizationName:         strings.TrimSpace(os.Getenv("ORGANIZATION_NAME")),
		IgnoreIfBootstrapped:     envBool("IGNORE_IF_BOOTSTRAPPED", false),
		InfisicalEmail:           strings.TrimSpace(os.Getenv("INFISICAL_EMAIL")),
		InfisicalPassword:        strings.TrimSpace(os.Getenv("INFISICAL_PASSWORD")),
		BootstrapSecretNamespace: strings.TrimSpace(os.Getenv("BOOTSTRAP_SECRET_NAMESPACE")),
		BootstrapSecretName:      strings.TrimSpace(os.Getenv("BOOTSTRAP_SECRET_NAME")),
		BootstrapEmailKey:        strings.TrimSpace(os.Getenv("BOOTSTRAP_SECRET_EMAIL_KEY")),
		BootstrapPasswordKey:     strings.TrimSpace(os.Getenv("BOOTSTRAP_SECRET_PASSWORD_KEY")),
		OrganizationID:           strings.TrimSpace(os.Getenv("ORGANIZATION_ID")),
		ProjectName:              strings.TrimSpace(os.Getenv("PROJECT_NAME")),
		ProjectSlug:              strings.TrimSpace(os.Getenv("PROJECT_SLUG")),
		EnvironmentName:          strings.TrimSpace(os.Getenv("ENVIRONMENT_NAME")),
		EnvironmentSlug:          strings.TrimSpace(os.Getenv("ENVIRONMENT_SLUG")),
		IdentityName:             strings.TrimSpace(os.Getenv("IDENTITY_NAME")),
		OrganizationIdentityRole: strings.TrimSpace(os.Getenv("ORGANIZATION_IDENTITY_ROLE")),
		IdentityRole:             strings.TrimSpace(os.Getenv("IDENTITY_ROLE")),
		EnableKubernetesAuth:     envBool("ENABLE_KUBERNETES_AUTH", false),
		WriteKubernetesSecret:    envBool("WRITE_KUBERNETES_SECRET", false),
		OutputStatusConfigMap:    strings.TrimSpace(os.Getenv("OUTPUT_STATUS_CONFIGMAP")),
		OutputProjectSecretName:  strings.TrimSpace(os.Getenv("OUTPUT_PROJECT_SECRET_NAME")),
		OutputProjectSecretKey:   strings.TrimSpace(os.Getenv("OUTPUT_PROJECT_SECRET_KEY")),
		SmokeTestSecretKey:       strings.TrimSpace(os.Getenv("SMOKE_TEST_SECRET_KEY")),
		SmokeTestSecretValue:     strings.TrimSpace(os.Getenv("SMOKE_TEST_SECRET_VALUE")),
		Secrets:                  loadSecrets("SECRETS_JSON"),
	}

	cfg.validate()

	return cfg
}

func loadMode() Mode {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BOOTSTRAP_MODE"))) {
	case "", string(ModeProject):
		return ModeProject
	case string(ModeOperator):
		return ModeOperator
	case string(ModePlatform):
		return ModePlatform
	default:
		panic(fmt.Sprintf("invalid BOOTSTRAP_MODE: %s", os.Getenv("BOOTSTRAP_MODE")))
	}
}

func (cfg *Config) validate() {
	switch cfg.Mode {
	case ModePlatform:
		cfg.requirePlatformFields()
	case ModeOperator:
		cfg.requireOperatorFields()
	case ModeProject:
		cfg.requireProjectFields()
	default:
		panic(fmt.Sprintf("unsupported bootstrap mode: %s", cfg.Mode))
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
}

func (cfg *Config) requirePlatformFields() {
	cfg.BootstrapEmail = mustEnv("BOOTSTRAP_EMAIL")
	cfg.BootstrapPassword = mustEnv("BOOTSTRAP_PASSWORD")
	cfg.OrganizationName = mustEnv("ORGANIZATION_NAME")
}

func (cfg *Config) requireProjectFields() {
	if cfg.BootstrapSecretNamespace != "" || cfg.BootstrapSecretName != "" {
		cfg.BootstrapSecretNamespace = mustEnv("BOOTSTRAP_SECRET_NAMESPACE")
		cfg.BootstrapSecretName = mustEnv("BOOTSTRAP_SECRET_NAME")
		if cfg.BootstrapEmailKey == "" {
			cfg.BootstrapEmailKey = "email"
		}
		if cfg.BootstrapPasswordKey == "" {
			cfg.BootstrapPasswordKey = "password"
		}
	} else {
		cfg.InfisicalEmail = mustEnv("INFISICAL_EMAIL")
		cfg.InfisicalPassword = mustEnv("INFISICAL_PASSWORD")
	}
	cfg.OrganizationID = mustEnv("ORGANIZATION_ID")
	cfg.ProjectName = mustEnv("PROJECT_NAME")
	cfg.ProjectSlug = mustEnv("PROJECT_SLUG")
	cfg.EnvironmentName = mustEnv("ENVIRONMENT_NAME")
	cfg.EnvironmentSlug = mustEnv("ENVIRONMENT_SLUG")
	cfg.IdentityName = mustEnv("IDENTITY_NAME")
	if cfg.IdentityRole == "" {
		cfg.IdentityRole = "member"
	}
}

func (cfg *Config) requireOperatorFields() {
	cfg.requireProjectFields()
	cfg.OrganizationIdentityRole = mustEnv("ORGANIZATION_IDENTITY_ROLE")
	cfg.EnableKubernetesAuth = true
	cfg.WriteKubernetesSecret = true
	cfg.KubernetesAuthHost = mustEnv("KUBERNETES_AUTH_HOST")
	cfg.AllowedNamespaces = mustEnv("ALLOWED_NAMESPACES")
	cfg.AllowedServiceAccounts = mustEnv("ALLOWED_SERVICE_ACCOUNTS")
	cfg.OutputSecretNamespace = mustEnv("OUTPUT_SECRET_NAMESPACE")
	cfg.OutputSecretName = mustEnv("OUTPUT_SECRET_NAME")
	cfg.OutputSecretKey = mustEnv("OUTPUT_SECRET_KEY")
	cfg.OutputProjectSecretName = mustEnv("OUTPUT_PROJECT_SECRET_NAME")
	cfg.OutputProjectSecretKey = mustEnv("OUTPUT_PROJECT_SECRET_KEY")
}

func loadSecrets(name string) []SecretSpec {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return nil
	}

	var arraySecrets []SecretSpec
	if err := json.Unmarshal([]byte(value), &arraySecrets); err != nil {
		panic(fmt.Sprintf("invalid JSON in %s: expected a JSON array of {key,value,path} objects: %v", name, err))
	}
	return normalizeSecretSpecs(name, arraySecrets)
}

func normalizeSecretSpecs(name string, specs []SecretSpec) []SecretSpec {
	cleaned := make([]SecretSpec, 0, len(specs))
	for _, spec := range specs {
		spec.Key = strings.TrimSpace(spec.Key)
		if spec.Key == "" {
			panic(fmt.Sprintf("%s contains an empty secret key", name))
		}
		spec.Path = normalizeSecretPath(spec.Path)
		cleaned = append(cleaned, spec)
	}
	return cleaned
}

func normalizeSecretPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
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
