package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	mockProviderPerms = 0644
)

type GitPluginProvider interface {
	PullGitHubPlugin(repoPath, repoUrl string) error
}

type GitPluginProviderImpl struct{}

func NewGitPluginProviderImpl() *GitPluginProviderImpl {
	return &GitPluginProviderImpl{}
}

func (gpp *GitPluginProviderImpl) PullGitHubPlugin(repoPath, repoUrl string) error {
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repoUrl, repoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone failed: %v\nOutput: %s", err, output)
		}
	} else {
		// If the repository already exists, pull the latest changes
		cmd := exec.Command("git", "-C", repoPath, "pull")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git pull failed: %v\nOutput: %s", err, output)
		}
	}
	return nil
}

type MockGitPluginProvider struct {
	// repoURL -> [ filename: fileContents ]
	github map[string]map[string]string
}

func NewMockGitPluginProvider(github map[string]map[string]string) *MockGitPluginProvider {
	return &MockGitPluginProvider{
		github: github,
	}
}

func (mgpp *MockGitPluginProvider) PullGitHubPlugin(repoPath, repoUrl string) error {
	repoContents, found := mgpp.github[repoUrl]
	if !found {
		return fmt.Errorf("Repo with url '%v' not found in github", repoUrl)
	}
	// repoPath should already exist but in case, create it
	err := os.MkdirAll(repoPath, 0744)
	if err != nil {
		return fmt.Errorf("An error occurred ensuring directory for '%v' exists:\n%v", repoUrl, err.Error())
	}

	for filename, contents := range repoContents {
		filePath := filepath.Join(repoPath, filename)

		err := os.WriteFile(filePath, []byte(contents), mockProviderPerms)
		if err != nil {
			return fmt.Errorf("An error occurred writing to filepath '%v' with contents:\n%v", repoPath, contents)
		}
	}
	return nil
}
