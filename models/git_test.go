package models_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitClient(t *testing.T) {
	tests := []struct {
		description   string
		common        models.CommonConfig
		disableGitLFS bool
		expectedErr   string
	}{
		{
			description: "create Git client with access token",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			disableGitLFS: false,
			expectedErr:   "",
		},
		{
			description: "create Git client with access token and skip SSL verification",
			common: models.CommonConfig{
				AccessToken:         "test-token",
				SkipSSLVerification: true,
			},
			disableGitLFS: false,
			expectedErr:   "",
		},
		{
			description: "create Git client with disabled Git LFS",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			disableGitLFS: true,
			expectedErr:   "",
		},
		{
			description: "create Git client with all options enabled",
			common: models.CommonConfig{
				AccessToken:         "test-token",
				SkipSSLVerification: true,
			},
			disableGitLFS: true,
			expectedErr:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tmpDir := t.TempDir()
			output := &bytes.Buffer{}

			client, err := models.NewGitClient(tc.common, tc.disableGitLFS, tmpDir, output)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, "test-token", client.AccessToken)
				assert.Equal(t, tmpDir, client.Directory)
				assert.Equal(t, output, client.Output)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, client)
			}
		})
	}
}

func TestNewGitClientWithApp(t *testing.T) {
	// Create a temporary private key file for testing
	tmpDir := t.TempDir()
	privateKeyPath := filepath.Join(tmpDir, "private.key")

	// Write a dummy private key (not valid, just for testing path handling)
	err := os.WriteFile(privateKeyPath, []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a..."), 0600)
	require.NoError(t, err)

	tests := []struct {
		description   string
		common        models.CommonConfig
		disableGitLFS bool
		expectedErr   string
	}{
		{
			description: "create Git client with GitHub App using private key",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			disableGitLFS: false,
			expectedErr:   "failed to create GitHub App transport",
		},
		{
			description: "create Git client with GitHub App using private key path",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKeyPath: privateKeyPath,
			},
			disableGitLFS: false,
			expectedErr:   "failed to create GitHub App transport",
		},
		{
			description: "create Git client with GitHub App and disabled Git LFS",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			disableGitLFS: true,
			expectedErr:   "failed to create GitHub App transport",
		},
		{
			description: "create Git client with GitHub App and skip SSL verification",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
				SkipSSLVerification:     true,
			},
			disableGitLFS: false,
			expectedErr:   "failed to create GitHub App transport",
		},
		{
			description: "create Git client with GitHub App with invalid app ID",
			common: models.CommonConfig{
				GithubAppID:             "not-a-number",
				GithubAppInstallationID: "67890",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			disableGitLFS: false,
			expectedErr:   "github_app_id",
		},
		{
			description: "create Git client with GitHub App with invalid installation ID",
			common: models.CommonConfig{
				GithubAppID:             "12345",
				GithubAppInstallationID: "not-a-number",
				GithubAppPrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a...",
			},
			disableGitLFS: false,
			expectedErr:   "github_app_installation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			workDir := t.TempDir()
			output := &bytes.Buffer{}

			client, err := models.NewGitClient(tc.common, tc.disableGitLFS, workDir, output)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, workDir, client.Directory)
				assert.Equal(t, output, client.Output)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, client)
			}
		})
	}
}

func TestNewGitClientRouting(t *testing.T) {
	tests := []struct {
		description      string
		common           models.CommonConfig
		expectedAuthType string
	}{
		{
			description: "should route to standard token auth when no GitHub App ID",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			expectedAuthType: "token",
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
			tmpDir := t.TempDir()
			output := &bytes.Buffer{}

			client, err := models.NewGitClient(tc.common, false, tmpDir, output)

			// For token auth: should succeed
			// For GitHub App: will error due to invalid key, but that's expected
			if tc.expectedAuthType == "token" {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, "test-token", client.AccessToken)
			} else {
				// GitHub App path will error due to invalid private key
				assert.Error(t, err)
			}
		})
	}
}

func TestNewGitClientEnvironmentVariables(t *testing.T) {
	tmpDir := t.TempDir()
	output := &bytes.Buffer{}

	// Save original environment variables
	originalSSLNoVerify := os.Getenv("GIT_SSL_NO_VERIFY")
	originalLFSSkipSmudge := os.Getenv("GIT_LFS_SKIP_SMUDGE")
	defer func() {
		os.Setenv("GIT_SSL_NO_VERIFY", originalSSLNoVerify)
		os.Setenv("GIT_LFS_SKIP_SMUDGE", originalLFSSkipSmudge)
	}()

	tests := []struct {
		description         string
		common              models.CommonConfig
		disableGitLFS       bool
		expectSSLNoVerify   bool
		expectLFSSkipSmudge bool
	}{
		{
			description: "no environment variables set when skip SSL and disable LFS are false",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			disableGitLFS:       false,
			expectSSLNoVerify:   false,
			expectLFSSkipSmudge: false,
		},
		{
			description: "GIT_SSL_NO_VERIFY set when skip SSL is true",
			common: models.CommonConfig{
				AccessToken:         "test-token",
				SkipSSLVerification: true,
			},
			disableGitLFS:       false,
			expectSSLNoVerify:   true,
			expectLFSSkipSmudge: false,
		},
		{
			description: "GIT_LFS_SKIP_SMUDGE set when disable LFS is true",
			common: models.CommonConfig{
				AccessToken: "test-token",
			},
			disableGitLFS:       true,
			expectSSLNoVerify:   false,
			expectLFSSkipSmudge: true,
		},
		{
			description: "both environment variables set when both options are true",
			common: models.CommonConfig{
				AccessToken:         "test-token",
				SkipSSLVerification: true,
			},
			disableGitLFS:       true,
			expectSSLNoVerify:   true,
			expectLFSSkipSmudge: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("GIT_SSL_NO_VERIFY")
			os.Unsetenv("GIT_LFS_SKIP_SMUDGE")

			_, err := models.NewGitClient(tc.common, tc.disableGitLFS, tmpDir, output)
			assert.NoError(t, err)

			if tc.expectSSLNoVerify {
				assert.Equal(t, "true", os.Getenv("GIT_SSL_NO_VERIFY"))
			} else {
				assert.Equal(t, "", os.Getenv("GIT_SSL_NO_VERIFY"))
			}

			if tc.expectLFSSkipSmudge {
				assert.Equal(t, "true", os.Getenv("GIT_LFS_SKIP_SMUDGE"))
			} else {
				assert.Equal(t, "", os.Getenv("GIT_LFS_SKIP_SMUDGE"))
			}
		})
	}
}
