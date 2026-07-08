package bootstrap

import (
	"fmt"
	"net/http"
	"net/url"
)

func ListProjects(c *HTTPClient, infisicalURL string, headers map[string]string) ([]Project, error) {
	payload, err := c.JSONRequest(http.MethodGet, infisicalURL+"/api/v1/projects", headers, nil)
	if err != nil {
		return nil, err
	}
	var resp ProjectsResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return nil, err
	}
	return resp.Projects, nil
}

func LoginWithPassword(c *HTTPClient, infisicalURL, email, password string) (string, error) {
	payload, err := c.JSONRequest(http.MethodPost, infisicalURL+"/api/v3/auth/login", map[string]string{
		"Accept":     "application/json",
		"User-Agent": "infisical-bootstrap-job",
	}, mustMarshal(map[string]string{
		"email":    email,
		"password": password,
	}))
	if err != nil {
		return "", err
	}
	var resp LoginResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}

func BootstrapInstance(c *HTTPClient, infisicalURL string, req BootstrapInstanceRequest) (BootstrapInstanceResponse, error) {
	payload, err := c.JSONRequest(http.MethodPost, infisicalURL+"/api/v1/admin/bootstrap", map[string]string{
		"Accept":     "application/json",
		"User-Agent": "infisical-bootstrap-job",
	}, mustMarshal(req))
	if err != nil {
		return BootstrapInstanceResponse{}, err
	}

	var resp BootstrapInstanceResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return BootstrapInstanceResponse{}, err
	}
	return resp, nil
}

func SelectOrganization(c *HTTPClient, infisicalURL, accessToken, organizationID string) (string, error) {
	payload, err := c.JSONRequest(http.MethodPost, infisicalURL+"/api/v3/auth/select-organization", map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
		"User-Agent":    "infisical-bootstrap-job",
	}, mustMarshal(map[string]string{
		"organizationId": organizationID,
		"userAgent":      "cli",
	}))
	if err != nil {
		return "", err
	}
	var resp SelectOrgResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return "", err
	}
	return resp.Token, nil
}

func EnsureProject(c *HTTPClient, infisicalURL string, headers map[string]string, projectName, projectSlug string) (Project, error) {
	projects, err := ListProjects(c, infisicalURL, headers)
	if err != nil {
		return Project{}, err
	}
	for _, item := range projects {
		if item.Slug == projectSlug {
			return item, nil
		}
	}

	payload, err := c.JSONRequest(http.MethodPost, infisicalURL+"/api/v1/projects", headers, mustMarshal(map[string]any{
		"projectName":             projectName,
		"slug":                    projectSlug,
		"type":                    "secret-manager",
		"shouldCreateDefaultEnvs": false,
		"hasDeleteProtection":     false,
	}))
	if err == nil {
		var createResp struct {
			Project Project `json:"project"`
		}
		if err := unmarshalInto(payload, &createResp); err != nil {
			return Project{}, err
		}
		return createResp.Project, nil
	}

	projects, retryErr := ListProjects(c, infisicalURL, headers)
	if retryErr != nil {
		return Project{}, err
	}
	for _, item := range projects {
		if item.Slug == projectSlug {
			return item, nil
		}
	}
	return Project{}, err
}

func EnsureEnvironment(c *HTTPClient, infisicalURL string, headers map[string]string, proj Project, envName, envSlug string) error {
	for _, item := range proj.Environments {
		if item.Slug == envSlug {
			return nil
		}
	}

	_, err := c.JSONRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/projects/%s/environments", infisicalURL, proj.ID), headers, mustMarshal(map[string]any{
		"name":     envName,
		"slug":     envSlug,
		"position": 1,
	}))
	if err == nil {
		return nil
	}

	refreshed, refreshErr := EnsureProject(c, infisicalURL, headers, proj.Name, proj.Slug)
	if refreshErr != nil {
		return err
	}
	for _, item := range refreshed.Environments {
		if item.Slug == envSlug {
			return nil
		}
	}
	return err
}

func ListIdentities(c *HTTPClient, infisicalURL string, headers map[string]string, orgID string) ([]IdentityRecord, error) {
	payload, err := c.JSONRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/identities?orgId=%s&offset=0&limit=1000", infisicalURL, url.QueryEscape(orgID)), headers, nil)
	if err != nil {
		return nil, err
	}
	var resp IdentitiesResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return nil, err
	}
	return resp.Identities, nil
}

func EnsureIdentity(c *HTTPClient, infisicalURL string, headers map[string]string, orgID, identityName, orgRole string) (string, string, error) {
	identities, err := ListIdentities(c, infisicalURL, headers, orgID)
	if err != nil {
		return "", "", err
	}
	for _, item := range identities {
		if item.Identity.Name == identityName {
			if orgRole != "" && item.Identity.Role != orgRole {
				if err := UpdateIdentityRole(c, infisicalURL, headers, item.IdentityID, orgRole); err != nil {
					return "", "", err
				}
			}
			return item.IdentityID, item.Identity.Name, nil
		}
	}

	if orgRole == "" {
		orgRole = "no-access"
	}

	payload, err := c.JSONRequest(http.MethodPost, infisicalURL+"/api/v1/identities", headers, mustMarshal(map[string]any{
		"name":                identityName,
		"organizationId":      orgID,
		"role":                orgRole,
		"hasDeleteProtection": false,
	}))
	if err == nil {
		var createResp CreateIdentityResponse
		if err := unmarshalInto(payload, &createResp); err != nil {
			return "", "", err
		}
		return createResp.Identity.ID, createResp.Identity.Name, nil
	}

	identities, retryErr := ListIdentities(c, infisicalURL, headers, orgID)
	if retryErr != nil {
		return "", "", err
	}
	for _, item := range identities {
		if item.Identity.Name == identityName {
			if orgRole != "" && item.Identity.Role != orgRole {
				if err := UpdateIdentityRole(c, infisicalURL, headers, item.IdentityID, orgRole); err != nil {
					return "", "", err
				}
			}
			return item.IdentityID, item.Identity.Name, nil
		}
	}
	return "", "", err
}

func UpdateIdentityRole(c *HTTPClient, infisicalURL string, headers map[string]string, identityID, orgRole string) error {
	_, err := c.JSONRequest(http.MethodPatch, fmt.Sprintf("%s/api/v1/identities/%s", infisicalURL, identityID), headers, mustMarshal(map[string]any{
		"role": orgRole,
	}))
	return err
}

func EnsureIdentityMembership(c *HTTPClient, infisicalURL string, headers map[string]string, projectID, identityID, identityRole string) error {
	listURL := fmt.Sprintf("%s/api/v1/projects/%s/memberships/identities?offset=0&limit=1000", infisicalURL, projectID)
	payload, err := c.JSONRequest(http.MethodGet, listURL, headers, nil)
	if err != nil {
		return err
	}
	var memberships MembershipsResponse
	if err := unmarshalInto(payload, &memberships); err != nil {
		return err
	}

	for _, item := range memberships.IdentityMemberships {
		if item.IdentityID != identityID {
			continue
		}
		if len(item.Roles) == 1 && item.Roles[0].Role == identityRole {
			return nil
		}
		_, err := c.JSONRequest(http.MethodPatch, fmt.Sprintf("%s/api/v1/projects/%s/memberships/identities/%s", infisicalURL, projectID, identityID), headers, mustMarshal(map[string]any{
			"roles": []map[string]any{{"role": identityRole, "isTemporary": false}},
		}))
		return err
	}

	_, err = c.JSONRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/projects/%s/memberships/identities/%s", infisicalURL, projectID, identityID), headers, mustMarshal(map[string]any{
		"roles": []map[string]any{{"role": identityRole, "isTemporary": false}},
	}))
	if err == nil {
		return nil
	}

	payload, retryErr := c.JSONRequest(http.MethodGet, listURL, headers, nil)
	if retryErr != nil {
		return err
	}
	if err := unmarshalInto(payload, &memberships); err != nil {
		return err
	}
	for _, item := range memberships.IdentityMemberships {
		if item.IdentityID == identityID {
			return nil
		}
	}
	return err
}

func EnsureKubernetesAuth(c *HTTPClient, infisicalURL string, headers map[string]string, identityID, kubernetesHost, caCert, allowedNamespaces, allowedNames string) error {
	body := mustMarshal(map[string]any{
		"kubernetesHost":          kubernetesHost,
		"caCert":                  caCert,
		"verifyTlsCertificate":    true,
		"tokenReviewMode":         "api",
		"allowedNamespaces":       allowedNamespaces,
		"allowedNames":            allowedNames,
		"allowedAudience":         "",
		"accessTokenTTL":          2592000,
		"accessTokenMaxTTL":       2592000,
		"accessTokenNumUsesLimit": 0,
		"accessTokenTrustedIps": []map[string]string{
			{"ipAddress": "0.0.0.0/0"},
			{"ipAddress": "::/0"},
		},
	})

	getURL := fmt.Sprintf("%s/api/v1/auth/kubernetes-auth/identities/%s", infisicalURL, identityID)
	if _, err := c.JSONRequest(http.MethodGet, getURL, headers, nil); err == nil {
		_, err := c.JSONRequest(http.MethodPatch, getURL, headers, body)
		return err
	}
	_, err := c.JSONRequest(http.MethodPost, getURL, headers, body)
	return err
}

func EnsureUniversalAuth(c *HTTPClient, infisicalURL string, headers map[string]string, identityID string) (string, error) {
	getURL := fmt.Sprintf("%s/api/v1/auth/universal-auth/identities/%s", infisicalURL, identityID)
	payload, err := c.JSONRequest(http.MethodGet, getURL, headers, nil)
	if err == nil {
		var resp UniversalAuthConfigResponse
		if err := unmarshalInto(payload, &resp); err != nil {
			return "", err
		}
		return resp.IdentityUniversalAuth.ClientID, nil
	}

	body := mustMarshal(map[string]any{
		"clientSecretTrustedIps": []map[string]string{
			{"ipAddress": "0.0.0.0/0"},
			{"ipAddress": "::/0"},
		},
		"accessTokenTrustedIps": []map[string]string{
			{"ipAddress": "0.0.0.0/0"},
			{"ipAddress": "::/0"},
		},
		"accessTokenTTL":          2592000,
		"accessTokenMaxTTL":       2592000,
		"accessTokenNumUsesLimit": 0,
	})
	payload, err = c.JSONRequest(http.MethodPost, getURL, headers, body)
	if err != nil {
		return "", err
	}

	var resp UniversalAuthConfigResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return "", err
	}
	return resp.IdentityUniversalAuth.ClientID, nil
}

func CreateUniversalAuthClientSecret(c *HTTPClient, infisicalURL string, headers map[string]string, identityID string) (string, error) {
	payload, err := c.JSONRequest(
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/auth/universal-auth/identities/%s/client-secrets", infisicalURL, identityID),
		headers,
		mustMarshal(map[string]any{
			"description":  "",
			"numUsesLimit": 0,
			"ttl":          0,
		}),
	)
	if err != nil {
		return "", err
	}

	var resp CreateUniversalAuthClientSecretResponse
	if err := unmarshalInto(payload, &resp); err != nil {
		return "", err
	}
	return resp.ClientSecret, nil
}

func EnsureSecretValue(c *HTTPClient, infisicalURL string, headers map[string]string, projectID, environmentSlug, secretKey, secretValue, secretPath string) error {
	secretPath = normalizeSecretPath(secretPath)
	getURL := fmt.Sprintf("%s/api/v4/secrets/%s?projectId=%s&environment=%s&secretPath=%s",
		infisicalURL,
		url.PathEscape(secretKey),
		url.QueryEscape(projectID),
		url.QueryEscape(environmentSlug),
		url.QueryEscape(secretPath),
	)
	if _, err := c.JSONRequest(http.MethodGet, getURL, headers, nil); err == nil {
		_, err := c.JSONRequest(http.MethodPatch, fmt.Sprintf("%s/api/v4/secrets/%s", infisicalURL, url.PathEscape(secretKey)), headers, mustMarshal(map[string]any{
			"projectId":   projectID,
			"environment": environmentSlug,
			"secretPath":  secretPath,
			"secretValue": secretValue,
			"type":        "shared",
		}))
		return err
	}
	_, err := c.JSONRequest(http.MethodPost, fmt.Sprintf("%s/api/v4/secrets/%s", infisicalURL, url.PathEscape(secretKey)), headers, mustMarshal(map[string]any{
		"projectId":   projectID,
		"environment": environmentSlug,
		"secretPath":  secretPath,
		"secretKey":   secretKey,
		"secretValue": secretValue,
		"type":        "shared",
	}))
	return err
}
