package promotion

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rhd-gitops-example/services/pkg/util"
)

type EnvLocation struct {
	RepoPath string // URL or local path
	Branch   string
	Folder   string
}

func (env EnvLocation) IsLocal() (bool, error) {
	parsed, err := url.Parse(env.RepoPath)
	if err != nil {
		return false, fmt.Errorf("failed parsing URL for environment path: %w", err)
	}
	return parsed.Scheme == "", nil
}

// String implements Stringer so that messages (including logs, commits and PRs)
// can reference EnvLocations consistently and in a more readable way.
// When possible, it will shorten repository paths so that 'github.com/org/repo'
// is just 'repo'. When not possible, it will use the entire URL.
func (env EnvLocation) String() string {
	var repoName string
	local, err := env.IsLocal()
	if err != nil {
		repoName = env.RepoPath
	} else {
		_, repoName, err = util.ExtractUserAndRepo(env.RepoPath)
		if err != nil {
			repoName = env.RepoPath
		}
	}
	if local {
		return "local filesystem directory"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "branch %s in %s", env.Branch, repoName)
	if env.Folder != "" {
		fmt.Fprintf(&b, " (environment folder %s)", env.Folder)
	}
	return b.String()
}
