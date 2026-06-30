package bootstrap

type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Slug         string        `json:"slug"`
	Environments []Environment `json:"environments"`
}

type Environment struct {
	Slug string `json:"slug"`
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type SelectOrgResponse struct {
	Token string `json:"token"`
}

type IdentitiesResponse struct {
	Identities []IdentityRecord `json:"identities"`
}

type IdentityRecord struct {
	IdentityID string `json:"identityId"`
	Identity   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"identity"`
}

type CreateIdentityResponse struct {
	Identity struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"identity"`
}

type MembershipsResponse struct {
	IdentityMemberships []Membership `json:"identityMemberships"`
}

type Membership struct {
	IdentityID string `json:"identityId"`
	Roles      []struct {
		Role string `json:"role"`
	} `json:"roles"`
}

type Result struct {
	ProjectID       string `json:"projectId"`
	ProjectSlug     string `json:"projectSlug"`
	IdentityID      string `json:"identityId"`
	IdentityName    string `json:"identityName"`
	EnvironmentSlug string `json:"environmentSlug"`
}
