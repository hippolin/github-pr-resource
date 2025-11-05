package pr

import (
	"errors"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
)

// Source represents the configuration for the resource.
type Source struct {
	models.CommonConfig
	models.GithubConfig
	Number        int      `json:"number"`
	GitCryptKey   string   `json:"git_crypt_key"`
	DisableGitLFS bool     `json:"disable_git_lfs"`
	Paths         []string `json:"paths"`
	IgnorePaths   []string `json:"ignore_paths"`
	DisableCISkip bool     `json:"disable_ci_skip"`
}

// Validate the source configuration.
func (s *Source) Validate() error {
	// Check if there is at least one authentication method
	hasAccessToken := s.AccessToken != ""
	hasGithubApp := s.GithubAppID != "" && s.GithubAppInstallationID != "" &&
		(s.GithubAppPrivateKey != "" || s.GithubAppPrivateKeyPath != "")

	if !hasAccessToken && !hasGithubApp {
		return errors.New("either access_token or github app credentials (github_app_id, github_app_installation_id, and github_app_private_key/github_app_private_key_path) must be set")
	}

	if hasAccessToken && hasGithubApp {
		return errors.New("cannot use both access_token and github app credentials")
	}

	if s.Repository == "" {
		return errors.New("repository must be set")
	}

	isHostingEndpointEnabled := s.HostingEndpoint != ""
	isV3EndpointEnabled := s.V3Endpoint != ""
	isV4EndpointEnabled := s.V4Endpoint != ""
	if isHostingEndpointEnabled != isV3EndpointEnabled || isV3EndpointEnabled != isV4EndpointEnabled {
		return errors.New("if any of hosting_endpoint, v3_endpoint, or v4_endpoint are set, all of them must be set")
	}
	return nil
}

type Version struct {
	Ref string `json:"ref"`
}
