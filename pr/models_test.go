package pr_test

import (
	"testing"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
	"github.com/cloudfoundry-community/github-pr-instances-resource/pr"
	"github.com/stretchr/testify/assert"
)

func TestSourceValidateAuthentication(t *testing.T) {
	tests := []struct {
		description string
		source      pr.Source
		expectedErr string
	}{
		{
			description: "validate succeeds with access token",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "",
		},
		{
			description: "validate succeeds with GitHub App using private key",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					GithubAppID:             "12345",
					GithubAppInstallationID: "67890",
					GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "",
		},
		{
			description: "validate succeeds with GitHub App using private key path",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					GithubAppID:             "12345",
					GithubAppInstallationID: "67890",
					GithubAppPrivateKeyPath: "/path/to/private.key",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "",
		},
		{
			description: "validate fails with neither access token nor GitHub App",
			source: pr.Source{
				CommonConfig: models.CommonConfig{},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "either access_token or github app credentials",
		},
		{
			description: "validate fails with both access token and GitHub App",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					AccessToken:             "my-token",
					GithubAppID:             "12345",
					GithubAppInstallationID: "67890",
					GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "cannot use both access_token and github app credentials",
		},
		{
			description: "validate fails with missing repository",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "",
				},
			},
			expectedErr: "repository must be set",
		},
		{
			description: "validate fails with GitHub App missing installation ID",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					GithubAppID:         "12345",
					GithubAppPrivateKey: "-----BEGIN RSA PRIVATE KEY-----",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "either access_token or github app credentials",
		},
		{
			description: "validate fails with GitHub App missing private key and path",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					GithubAppID:             "12345",
					GithubAppInstallationID: "67890",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "either access_token or github app credentials",
		},
		{
			description: "validate fails with mismatched endpoint configuration - only hosting endpoint",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository:      "owner/repo",
					HostingEndpoint: "https://github.enterprise.com",
				},
			},
			expectedErr: "if any of hosting_endpoint, v3_endpoint, or v4_endpoint are set, all of them must be set",
		},
		{
			description: "validate fails with mismatched endpoint configuration - only v3 endpoint",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
					V3Endpoint: "https://github.enterprise.com/api/v3",
				},
			},
			expectedErr: "if any of hosting_endpoint, v3_endpoint, or v4_endpoint are set, all of them must be set",
		},
		{
			description: "validate succeeds with all endpoints configured",
			source: pr.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository:      "owner/repo",
					HostingEndpoint: "https://github.enterprise.com",
					V3Endpoint:      "https://github.enterprise.com/api/v3",
					V4Endpoint:      "https://github.enterprise.com/api/graphql",
				},
			},
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.source.Validate()

			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
