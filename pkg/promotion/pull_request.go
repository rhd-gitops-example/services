package promotion

import (
	"github.com/jenkins-x/go-scm/scm"
)

// TODO: OptionFunc for Title?
// TODO: For the Head, should this try and determine whether or not this is a
// fork ("user" of both repoURLs) and if so, simplify the Head?
func makePullRequestInput(from, to EnvLocation, branchName, prBody string) (*scm.PullRequestInput, error) {
	return &scm.PullRequestInput{
		Title: prBody,
		Head:  branchName,
		Base:  to.Branch,
		Body:  prBody,
	}, nil
}
