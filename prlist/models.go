package prlist

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
	"github.com/shurcooL/githubv4"
)

// Source represents the configuration for the resource.
type Source struct {
	models.CommonConfig
	models.GithubConfig
	Paths                   []string                    `json:"paths"`
	IgnorePaths             []string                    `json:"ignore_paths"`
	DisableCISkip           bool                        `json:"disable_ci_skip"`
	DisableForks            bool                        `json:"disable_forks"`
	IgnoreDrafts            bool                        `json:"ignore_drafts"`
	BaseBranch              string                      `json:"base_branch"`
	RequiredReviewApprovals int                         `json:"required_review_approvals"`
	GitCryptKey             string                      `json:"git_crypt_key"`
	Labels                  []string                    `json:"labels"`
	States                  []githubv4.PullRequestState `json:"states"`
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
	if s.V3Endpoint != "" && s.V4Endpoint == "" {
		return errors.New("v4_endpoint must be set together with v3_endpoint")
	}
	if s.V4Endpoint != "" && s.V3Endpoint == "" {
		return errors.New("v3_endpoint must be set together with v4_endpoint")
	}
	for _, state := range s.States {
		switch state {
		case githubv4.PullRequestStateOpen:
		case githubv4.PullRequestStateClosed:
		case githubv4.PullRequestStateMerged:
		default:
			return fmt.Errorf("states value \"%s\" must be one of: OPEN, MERGED, CLOSED", state)
		}
	}
	return nil
}

type Version struct {
	// JSON encoded list of PR numbers.
	PRs string `json:"prs"`
	// Time when the version was initially generated.
	Timestamp string `json:"timestamp"`
}

// NewVersion constructs a new Version.
func NewVersion(prs []*models.PullRequest) Version {
	numbers := make([]int, len(prs))
	for i, pr := range prs {
		numbers[i] = pr.Number
	}
	data, err := json.Marshal(numbers)
	if err != nil {
		panic(err)
	}
	return Version{
		PRs:       string(data),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// PRData represents a single PR in the get response file.
type PRData struct {
	Number int `json:"number"`
}
