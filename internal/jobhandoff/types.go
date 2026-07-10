package jobhandoff

type Metadata struct {
	InfisicalURL              string `json:"infisicalUrl"`
	OrganizationID            string `json:"organizationId"`
	ProjectID                 string `json:"projectId"`
	ProjectSlug               string `json:"projectSlug"`
	EnvironmentSlug           string `json:"environmentSlug"`
	IdentityID                string `json:"identityId"`
	IdentityName              string `json:"identityName"`
	UniversalAuthClientID     string `json:"universalAuthClientId"`
	UniversalAuthClientSecret string `json:"universalAuthClientSecret"`
}
