package models

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
)

// Git interface for testing purposes.
//
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/fake_git.go . Git
type Git interface {
	Init(*string) error
	Pull(string, string, int, bool, bool) error
	RevParse(string) (string, error)
	RevList(*string, []string, []string, bool) ([]string, error)
	Fetch(string, int, int, bool, bool) error
	Checkout(string, string, bool) error
	Merge(string, bool) error
	Rebase(string, string, bool) error
	GitCryptUnlock(string) error
}

func NewGitClient(common CommonConfig, disableGitLFS bool, dir string, output io.Writer) (*GitClient, error) {
	// Check if GitHub App authentication should be used
	if common.GithubAppID != "" {
		return NewGitClientWithApp(common, disableGitLFS, dir, output)
	}

	if common.SkipSSLVerification {
		os.Setenv("GIT_SSL_NO_VERIFY", "true")
	}
	if disableGitLFS {
		os.Setenv("GIT_LFS_SKIP_SMUDGE", "true")
	}
	return &GitClient{
		AccessToken: common.AccessToken,
		Directory:   dir,
		Output:      output,
	}, nil
}

// NewGitClientWithApp creates a GitClient authenticated via GitHub App
func NewGitClientWithApp(common CommonConfig, disableGitLFS bool, dir string, output io.Writer) (*GitClient, error) {
	if disableGitLFS {
		os.Setenv("GIT_LFS_SKIP_SMUDGE", "true")
	}

	// Configure HTTP transport for GitHub App authentication
	var tr *http.Transport
	if common.SkipSSLVerification {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		os.Setenv("GIT_SSL_NO_VERIFY", "true")
	} else {
		tr = http.DefaultTransport.(*http.Transport)
	}

	githubAppID, err := toInt64(common.GithubAppID)
	if err != nil {
		return nil, fmt.Errorf("github_app_id: %v", err)
	}

	githubAppInstallationID, err := toInt64(common.GithubAppInstallationID)
	if err != nil {
		return nil, fmt.Errorf("github_app_installation: %v", err)
	}

	// Create a GitHub App transport using provided credentials
	var transport *ghinstallation.Transport
	// var err error
	if common.GithubAppPrivateKeyPath != "" {
		transport, err = ghinstallation.NewKeyFromFile(
			tr,
			githubAppID,
			githubAppInstallationID,
			common.GithubAppPrivateKeyPath,
		)
	} else {
		transport, err = ghinstallation.New(
			tr,
			githubAppID,
			githubAppInstallationID,
			[]byte(common.GithubAppPrivateKey),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub App transport: %v", err)
	}

	// Generate a token for Git authentication
	token, err := transport.Token(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub App token: %v", err)
	}

	return &GitClient{
		AccessToken: token,
		Directory:   dir,
		Output:      output,
		// Store the transport for potential token refresh if needed
		transport: transport,
	}, nil
}

// GitClient ...
type GitClient struct {
	AccessToken string
	Directory   string
	Output      io.Writer

	transport *ghinstallation.Transport // For GitHub App authentication
}

func (g *GitClient) silentCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = g.Directory
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		"X_OAUTH_BASIC_TOKEN="+g.AccessToken,
		"GIT_ASKPASS=/usr/local/bin/askpass.sh")
	return cmd
}

func (g *GitClient) command(name string, arg ...string) *exec.Cmd {
	cmd := g.silentCommand(name, arg...)
	cmd.Stdout = g.Output
	cmd.Stderr = g.Output
	return cmd
}

func (g *GitClient) Init(branch *string) error {
	if err := g.command("git", "init", "-b", "main").Run(); err != nil {
		return fmt.Errorf("init failed: %s", err)
	}
	if branch != nil {
		if err := g.command("git", "checkout", "-b", *branch).Run(); err != nil {
			return fmt.Errorf("checkout to '%s' failed: %s", *branch, err)
		}
	}
	if err := g.command("git", "config", "user.name", "concourse-ci").Run(); err != nil {
		return fmt.Errorf("failed to configure git user: %s", err)
	}
	if err := g.command("git", "config", "user.email", "concourse@local").Run(); err != nil {
		return fmt.Errorf("failed to configure git email: %s", err)
	}
	if err := g.command("git", "config", "url.https://x-oauth-basic@github.com/.insteadOf", "git@github.com:").Run(); err != nil {
		return fmt.Errorf("failed to configure github url: %s", err)
	}
	if err := g.command("git", "config", "url.https://.insteadOf", "git://").Run(); err != nil {
		return fmt.Errorf("failed to configure github url: %s", err)
	}
	return nil
}

// Pull ...
func (g *GitClient) Pull(uri, branch string, depth int, submodules bool, fetchTags bool) error {
	endpoint, err := g.Endpoint(uri)
	if err != nil {
		return err
	}

	if err := g.command("git", "remote", "add", "origin", endpoint).Run(); err != nil {
		return fmt.Errorf("setting 'origin' remote to '%s' failed: %s", endpoint, err)
	}

	args := []string{"pull", "origin", branch}
	if depth > 0 {
		args = append(args, "--depth", strconv.Itoa(depth))
	}
	if fetchTags {
		args = append(args, "--tags")
	}
	if submodules {
		args = append(args, "--recurse-submodules")
	}
	cmd := g.command("git", args...)

	// Discard output to have zero chance of logging the access token.
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pull failed: %s", cmd)
	}
	if submodules {
		submodulesGet := g.command("git", "submodule", "update", "--init", "--recursive")
		if err := submodulesGet.Run(); err != nil {
			return fmt.Errorf("submodule update failed: %s", err)
		}
	}
	return nil
}

// RevParse retrieves the SHA of the given branch.
func (g *GitClient) RevParse(branch string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = g.Directory
	sha, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("rev-parse '%s' failed: %s: %s", branch, err, string(sha))
	}
	return strings.TrimSpace(string(sha)), nil
}

// RevList retrieves the list of commits starting with (and including)
// fromCommit (if it exists) in chronological order. If fromCommit is empty or
// does not exist in the repo, only the latest commit will be returned.
//
// It also
func (g *GitClient) RevList(fromCommit *string, paths []string, ignorePaths []string, disableCISkip bool) ([]string, error) {
	missingFromCommit := fromCommit == nil || !g.commitExists(*fromCommit)

	initCommitBytes, err := g.silentCommand("git", "rev-list", "--max-parents=0", "HEAD").Output()
	if err != nil {
		return nil, err
	}
	initCommitsLines := strings.Split(strings.TrimSpace(string(initCommitBytes)), "\n")
	initCommit := initCommitsLines[len(initCommitsLines)-1]
	var logRange string
	if missingFromCommit || *fromCommit == initCommit {
		logRange = "HEAD"
	} else {
		logRange = *fromCommit + "~1..HEAD"
	}

	args := []string{
		"rev-list", "--first-parent", logRange, "--reverse",
	}

	if !disableCISkip {
		args = append(args, "--invert-grep", `--grep=\[skip\sci\]`, `--grep=\[ci\sskip\]`)
	}

	if missingFromCommit {
		// Only grab the latest commit if there is no (valid) from commit
		args = append(args, "-1")
	}

	if len(paths) > 0 || len(ignorePaths) > 0 {
		args = append(args, "--")
		if len(paths) > 0 {
			args = append(args, paths...)
		} else {
			args = append(args, ".")
		}

		ignorePathArgs := make([]string, len(ignorePaths))
		for i, ignorePath := range ignorePaths {
			ignorePathArgs[i] = ":!" + ignorePath
		}
		args = append(args, ignorePathArgs...)
	}

	output, err := g.silentCommand("git", args...).Output()
	if err != nil {
		return nil, err
	}
	commitLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return commitLines, nil
}

func (g *GitClient) commitExists(commit string) bool {
	err := g.command("git", "cat-file", "-e", commit).Run()
	return err == nil
}

func (g *GitClient) Fetch(uri string, prNumber int, depth int, submodules bool, checkout bool) error {
	endpoint, err := g.Endpoint(uri)
	if err != nil {
		return err
	}

	args := []string{"fetch", endpoint, fmt.Sprintf("pull/%s/head", strconv.Itoa(prNumber))}
	if depth > 0 {
		args = append(args, "--depth", strconv.Itoa(depth))
	}
	if submodules {
		args = append(args, "--recurse-submodules")
	}
	cmd := g.command("git", args...)

	// Discard output to have zero chance of logging the access token.
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fetch failed: %v", err)
	}

	if checkout {
		cmd := g.command("git", "checkout", "FETCH_HEAD")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("checkout failed: %v", err)
		}
	}

	return nil
}

// CheckOut
func (g *GitClient) Checkout(branch, sha string, submodules bool) error {
	if err := g.command("git", "checkout", "-b", branch, sha).Run(); err != nil {
		return fmt.Errorf("checkout failed: %s", err)
	}

	if submodules {
		if err := g.command("git", "submodule", "update", "--init", "--recursive", "--checkout").Run(); err != nil {
			return fmt.Errorf("submodule update failed: %s", err)
		}
	}

	return nil
}

// Merge ...
func (g *GitClient) Merge(sha string, submodules bool) error {
	if err := g.command("git", "merge", sha, "--no-stat").Run(); err != nil {
		return fmt.Errorf("merge failed: %s", err)
	}

	if submodules {
		if err := g.command("git", "submodule", "update", "--init", "--recursive", "--merge").Run(); err != nil {
			return fmt.Errorf("submodule update failed: %s", err)
		}
	}

	return nil
}

// Rebase ...
func (g *GitClient) Rebase(baseRef string, headSha string, submodules bool) error {
	if err := g.command("git", "rebase", baseRef, headSha).Run(); err != nil {
		return fmt.Errorf("rebase failed: %s", err)
	}

	if submodules {
		if err := g.command("git", "submodule", "update", "--init", "--recursive", "--rebase").Run(); err != nil {
			return fmt.Errorf("submodule update failed: %s", err)
		}
	}

	return nil
}

// GitCryptUnlock unlocks the repository using git-crypt
func (g *GitClient) GitCryptUnlock(base64key string) error {
	return fmt.Errorf("GitCrypt Unsupported")
}

// Endpoint takes an uri and produces an endpoint with the login information baked in.
func (g *GitClient) Endpoint(uri string) (string, error) {
	endpoint, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse commit url: %s", err)
	}

	// If using GitHub App authentication and token needs refresh
	if g.transport != nil {
		token, err := g.transport.Token(nil)
		if err != nil {
			return "", fmt.Errorf("failed to refresh GitHub App token: %v", err)
		}
		g.AccessToken = token
	}

	endpoint.User = url.UserPassword("x-oauth-basic", g.AccessToken)
	return endpoint.String(), nil
}
