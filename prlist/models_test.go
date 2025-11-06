package prlist_test

import (
	"testing"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
	"github.com/cloudfoundry-community/github-pr-instances-resource/prlist"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
)

func TestSourceValidateAuthentication(t *testing.T) {
	tests := []struct {
		description string
		source      prlist.Source
		expectedErr string
	}{
		{
			description: "validate succeeds with access token",
			source: prlist.Source{
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
			source: prlist.Source{
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
			source: prlist.Source{
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
			source: prlist.Source{
				CommonConfig: models.CommonConfig{},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
			},
			expectedErr: "either access_token or github app credentials",
		},
		{
			description: "validate fails with both access token and GitHub App",
			source: prlist.Source{
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
			source: prlist.Source{
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
			source: prlist.Source{
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
			source: prlist.Source{
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
			description: "validate fails with only v3 endpoint",
			source: prlist.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
					V3Endpoint: "https://github.enterprise.com/api/v3",
				},
			},
			expectedErr: "v4_endpoint must be set together with v3_endpoint",
		},
		{
			description: "validate fails with only v4 endpoint",
			source: prlist.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
					V4Endpoint: "https://github.enterprise.com/api/graphql",
				},
			},
			expectedErr: "v3_endpoint must be set together with v4_endpoint",
		},
		{
			description: "validate succeeds with both v3 and v4 endpoints",
			source: prlist.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
					V3Endpoint: "https://github.enterprise.com/api/v3",
					V4Endpoint: "https://github.enterprise.com/api/graphql",
				},
			},
			expectedErr: "",
		},
		{
			description: "validate succeeds with valid PR states",
			source: prlist.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
				States: []githubv4.PullRequestState{
					githubv4.PullRequestStateOpen,
					githubv4.PullRequestStateMerged,
					githubv4.PullRequestStateClosed,
				},
			},
			expectedErr: "",
		},
		{
			description: "validate fails with invalid PR state",
			source: prlist.Source{
				CommonConfig: models.CommonConfig{
					AccessToken: "my-token",
				},
				GithubConfig: models.GithubConfig{
					Repository: "owner/repo",
				},
				States: []githubv4.PullRequestState{
					githubv4.PullRequestStateOpen,
					"INVALID_STATE",
				},
			},
			expectedErr: "must be one of: OPEN, MERGED, CLOSED",
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
