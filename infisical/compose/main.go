package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"infisical-bootstrap-job/internal/bootstrap"
	"infisical-bootstrap-job/internal/jobhandoff"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "compose bootstrap failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	infisicalURL := jobhandoff.Env("INFISICAL_URL", "http://infisical:8080")
	outputPath := jobhandoff.Env("BOOTSTRAP_METADATA_PATH", "/bootstrap/agent-vault.json")

	if metadata, err := jobhandoff.ReadMetadata(outputPath); err == nil {
		fmt.Printf("Bootstrap metadata already exists at %s for project %s\n", outputPath, metadata.ProjectSlug)
		return nil
	}

	if err := waitForInfisical(infisicalURL); err != nil {
		return err
	}

	platformResp, err := runPlatform(infisicalURL)
	if err != nil {
		if bootstrap.IsAlreadyBootstrappedError(err) {
			return fmt.Errorf("infisical is already bootstrapped but %s is missing; keep the bootstrap volume or reset with docker compose down -v", outputPath)
		}
		return err
	}

	projectResp, err := runProject(infisicalURL, platformResp.Organization.ID)
	if err != nil {
		return err
	}

	metadata := jobhandoff.Metadata{
		InfisicalURL:              infisicalURL,
		OrganizationID:            platformResp.Organization.ID,
		ProjectID:                 projectResp.ProjectID,
		ProjectSlug:               projectResp.ProjectSlug,
		EnvironmentSlug:           projectResp.EnvironmentSlug,
		IdentityID:                projectResp.IdentityID,
		IdentityName:              projectResp.IdentityName,
		UniversalAuthClientID:     projectResp.UniversalAuthClientID,
		UniversalAuthClientSecret: projectResp.UniversalAuthClientSecret,
	}

	if err := jobhandoff.WriteMetadata(outputPath, metadata); err != nil {
		return err
	}

	fmt.Printf("Wrote bootstrap metadata to %s\n", outputPath)
	return nil
}

func waitForInfisical(infisicalURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(5 * time.Minute)

	for {
		req, err := http.NewRequest(http.MethodGet, infisicalURL+"/", nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return nil
			}
		}

		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("timed out waiting for Infisical at %s: %w", infisicalURL, err)
			}
			return fmt.Errorf("timed out waiting for Infisical at %s", infisicalURL)
		}
		time.Sleep(2 * time.Second)
	}
}

func runPlatform(infisicalURL string) (bootstrap.BootstrapInstanceResponse, error) {
	api := bootstrap.NewHTTPClient()
	return bootstrap.BootstrapInstance(api, infisicalURL, bootstrap.BootstrapInstanceRequest{
		Email:        jobhandoff.MustEnv("BOOTSTRAP_EMAIL"),
		Password:     jobhandoff.MustEnv("BOOTSTRAP_PASSWORD"),
		Organization: jobhandoff.MustEnv("ORGANIZATION_NAME"),
	})
}

func runProject(infisicalURL, organizationID string) (bootstrap.Result, error) {
	api := bootstrap.NewHTTPClient()
	email := jobhandoff.MustEnv("BOOTSTRAP_EMAIL")
	password := jobhandoff.MustEnv("BOOTSTRAP_PASSWORD")

	sessionToken, err := bootstrap.LoginWithPassword(api, infisicalURL, email, password)
	if err != nil {
		return bootstrap.Result{}, err
	}
	orgToken, err := bootstrap.SelectOrganization(api, infisicalURL, sessionToken, organizationID)
	if err != nil {
		return bootstrap.Result{}, err
	}
	headers := bootstrap.BearerHeaders(orgToken)

	project, err := bootstrap.EnsureProject(api, infisicalURL, headers, jobhandoff.MustEnv("INFISICAL_PROJECT_NAME"), jobhandoff.MustEnv("INFISICAL_PROJECT_SLUG"))
	if err != nil {
		return bootstrap.Result{}, err
	}
	if err := bootstrap.EnsureEnvironment(api, infisicalURL, headers, project, jobhandoff.MustEnv("INFISICAL_ENVIRONMENT_NAME"), jobhandoff.MustEnv("INFISICAL_ENVIRONMENT_SLUG")); err != nil {
		return bootstrap.Result{}, err
	}

	identityID, identityName, err := bootstrap.EnsureIdentity(api, infisicalURL, headers, organizationID, jobhandoff.MustEnv("INFISICAL_IDENTITY_NAME"), "")
	if err != nil {
		return bootstrap.Result{}, err
	}
	if err := bootstrap.EnsureIdentityMembership(api, infisicalURL, headers, project.ID, identityID, "member"); err != nil {
		return bootstrap.Result{}, err
	}

	clientID, err := bootstrap.EnsureUniversalAuth(api, infisicalURL, headers, identityID)
	if err != nil {
		return bootstrap.Result{}, err
	}
	clientSecret, err := bootstrap.CreateUniversalAuthClientSecret(api, infisicalURL, headers, identityID)
	if err != nil {
		return bootstrap.Result{}, err
	}

	secrets := bootstrap.LoadSecretsFromString(jobhandoff.MustEnv("INFISICAL_SECRETS_JSON"))
	for _, secret := range secrets {
		if err := bootstrap.EnsureSecretValue(api, infisicalURL, headers, project.ID, jobhandoff.MustEnv("INFISICAL_ENVIRONMENT_SLUG"), secret.Key, secret.Value, secret.Path); err != nil {
			return bootstrap.Result{}, err
		}
	}

	return bootstrap.Result{
		ProjectID:                 project.ID,
		ProjectSlug:               project.Slug,
		IdentityID:                identityID,
		IdentityName:              identityName,
		EnvironmentSlug:           jobhandoff.MustEnv("INFISICAL_ENVIRONMENT_SLUG"),
		UniversalAuthClientID:     clientID,
		UniversalAuthClientSecret: clientSecret,
	}, nil
}
