package mock

import "github.com/bigkevmcd/services/pkg/git"

var _ git.Cache = (*MockCache)(nil)
