package avancement

import (
	"fmt"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"

	"github.com/bigkevmcd/services/pkg/git/mock"
)

const (
	dev     = "https://example.com/testing/dev-env"
	staging = "https://example.com/testing/staging-env"
)

var testBody = []byte("this is the body")

func TestPromoteWithSuccess(t *testing.T) {
	filePath := pathForService("my-service")
	client, data := fakescm.NewDefault()
	fakeClientFactory := func(s string) *scm.Client {
		return client
	}

	mc := mock.New()
	mc.AddRepoFile(dev, "master", filePath, testBody)

	sm := New(mc, "testing")
	sm.clientFactory = fakeClientFactory

	err := sm.Promote("my-service", dev, staging, "my-branch")
	if err != nil {
		t.Fatal(err)
	}

	mc.AssertBranchCreated(t, staging, "master", "my-branch")
	mc.AssertFileWrittenToBranch(t, staging, "my-branch", filePath, testBody)
	mc.AssertCommitAndPush(t, staging, "my-branch", "this is a test commit", "testing")
}
