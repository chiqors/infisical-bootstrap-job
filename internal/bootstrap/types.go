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

type BootstrapInstanceRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Organization string `json:"organization"`
}

type BootstrapInstanceResponse struct {
	Message      string                `json:"message"`
	User         BootstrapUser         `json:"user"`
	Organization BootstrapOrganization `json:"organization"`
	Identity     BootstrapIdentity     `json:"identity"`
}

type BootstrapUser struct {
	Username   string `json:"username"`
	ID         string `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	SuperAdmin bool   `json:"superAdmin"`
}

type BootstrapOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type BootstrapIdentity struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Credentials BootstrapIdentityCredential `json:"credentials"`
}

type BootstrapIdentityCredential struct {
	Token string `json:"token"`
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
		Role string `json:"role"`
	} `json:"identity"`
}

type CreateIdentityResponse struct {
	Identity struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
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
