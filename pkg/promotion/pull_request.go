package promotion

import (
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/rhd-gitops-example/services/pkg/util"
)

// TODO: OptionFunc for Title?
// TODO: For the Head, should this try and determine whether or not this is a
// fork ("user" of both repoURLs) and if so, simplify the Head?
func makePullRequestInput(from, to EnvLocation, branchName, prBody string) (*scm.PullRequestInput, error) {
	var title string

	_, toRepo, err := util.ExtractUserAndRepo(to.RepoPath)
	if err != nil {
		return nil, err
	}

	if from.IsLocal() {
		title = fmt.Sprintf("promotion from local filesystem directory to %s", toRepo)
	} else {
		_, fromRepo, err := util.ExtractUserAndRepo(from.RepoPath)
		if err != nil {
			return nil, err
		}
		title = fmt.Sprintf("promotion from %s to %s", fromRepo, toRepo)
	}

	return &scm.PullRequestInput{
		Title: title,
		Head:  branchName,
		Base:  to.Branch,
		Body:  prBody,
	}, nil
}
