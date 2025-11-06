package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGithubClient(t *testing.T) {
	tests := []struct {
		description string
		common      models.CommonConfig
		config      models.GithubConfig
		expectedErr string
	}{
		{
			description: "create GitHub client with access token",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "",
		},
		{
			description: "create GitHub client with access token and skip SSL verification",
			common: models.CommonConfig{
				AccessToken:         "test-token",
				SkipSSLVerification: true,
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "",
		},
		{
			description: "create GitHub client with invalid repository format",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			config: models.GithubConfig{
				Repository: "invalid-repo-format",
			},
			expectedErr: "malformed repository",
		},
		{
			description: "create GitHub client with v3 and v4 endpoints",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
				V3Endpoint: "https://github.enterprise.com/api/v3",
				V4Endpoint: "https://github.enterprise.com/api/graphql",
			},
			expectedErr: "",
		},
		{
			description: "create GitHub client with invalid v3 endpoint URL",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
				V3Endpoint: "://invalid-url",
				V4Endpoint: "https://github.enterprise.com/api/graphql",
			},
			expectedErr: "failed to parse v3 endpoint",
		},
		{
			description: "create GitHub client with invalid v4 endpoint URL",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
				V3Endpoint: "https://github.enterprise.com/api/v3",
				V4Endpoint: "://invalid-url",
			},
			expectedErr: "failed to parse v4 endpoint",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			client, err := models.NewGithubClient(tc.common, tc.config)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.V3)
				assert.NotNil(t, client.V4)
				assert.Equal(t, "repo", client.Repository)
				assert.Equal(t, "owner", client.Owner)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, client)
			}
		})
	}
}

func TestNewGithubClientWithApp(t *testing.T) {
	// Create a temporary private key file for testing
	tmpDir := t.TempDir()
	privateKeyPath := filepath.Join(tmpDir, "private.key")

	// Write a dummy private key (this is not a valid key, just for testing path handling)
	err := os.WriteFile(privateKeyPath, []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a..."), 0600)
	require.NoError(t, err)

	tests := []struct {
		description string
		common      models.CommonConfig
		config      models.GithubConfig
		expectedErr string
	}{
		{
			description: "create GitHub client with GitHub App using private key",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "failed to create GitHub App transport",
		},
		{
			description: "create GitHub client with GitHub App using private key path",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKeyPath: privateKeyPath,
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "failed to create GitHub App transport",
		},
		{
			description: "create GitHub client with GitHub App with invalid app ID",
			common: models.CommonConfig{
				GithubAppID:             "not-a-number",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "github_app_id",
		},
		{
			description: "create GitHub client with GitHub App with invalid installation ID",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "not-a-number",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "github_app_installation",
		},
		{
			description: "create GitHub client with GitHub App and invalid repository",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "invalid-repo",
			},
			expectedErr: "malformed repository",
		},
		{
			description: "create GitHub client with GitHub App and skip SSL verification",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
				SkipSSLVerification:     true,
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
			},
			expectedErr: "failed to create GitHub App transport",
		},
		{
			description: "create GitHub client with GitHub App and custom v3 endpoint",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
				V3Endpoint: "https://github.enterprise.com/api/v3",
				V4Endpoint: "https://github.enterprise.com/api/graphql",
			},
			expectedErr: "failed to create GitHub App transport",
		},
		{
			description: "create GitHub client with GitHub App and invalid v3 endpoint",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
				V3Endpoint: "://invalid-url",
				V4Endpoint: "https://github.enterprise.com/api/graphql",
			},
			expectedErr: "failed to create GitHub App transport",
		},
		{
			description: "create GitHub client with GitHub App and invalid v4 endpoint",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			config: models.GithubConfig{
				Repository: "owner/repo",
				V3Endpoint: "https://github.enterprise.com/api/v3",
				V4Endpoint: "://invalid-url",
			},
			expectedErr: "failed to create GitHub App transport",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			client, err := models.NewGithubClient(tc.common, tc.config)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.V3)
				assert.NotNil(t, client.V4)
				assert.Equal(t, "repo", client.Repository)
				assert.Equal(t, "owner", client.Owner)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, client)
			}
		})
	}
}

func TestNewGithubClientRouting(t *testing.T) {
	tests := []struct {
		description      string
		common           models.CommonConfig
		expectedAuthType string
	}{
		{
			description: "should route to standard OAuth2 when no GitHub App ID",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			expectedAuthType: "oauth2",
		},
		{
			description: "should route to GitHub App when GitHub App ID is set",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			expectedAuthType: "github-app",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			config := models.GithubConfig{
				Repository: "owner/repo",
			}

			client, err := models.NewGithubClient(tc.common, config)

			// For OAuth2: should succeed with valid config
			// For GitHub App: will error due to invalid key, but that's expected
			if tc.expectedAuthType == "oauth2" {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			} else {
				// GitHub App path will error due to invalid private key
				assert.Error(t, err)
			}
		})
	}
}
