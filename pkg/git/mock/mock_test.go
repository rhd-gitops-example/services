package mock

import "github.com/rhd-gitops-example/services/pkg/git"

var _ git.Repo = (*Repository)(nil)
