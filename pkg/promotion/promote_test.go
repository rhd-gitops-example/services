package promotion

import (
	"testing"

	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/git/mock"
)

func TestGetEnvironmentFolder(t *testing.T) {
	noFolders := mock.New("nonePath", "master")
	oneFolder := mock.New("oenPath", "master")
	oneFolder.AddFiles("example")
	twoFolders := mock.New("twoPath", "master")
	twoFolders.AddFiles("example")
	twoFolders.AddFiles("another")

	tests := []struct {
		repo           git.Repo
		checkFolder    string
		expectedFolder string
		expectedError  bool
	}{
		{noFolders, "", "", true},
		{noFolders, "example", "", true},
		{oneFolder, "", "example", false},
		{oneFolder, "example", "example", false},
		{oneFolder, "another", "", true},
		{twoFolders, "", "", true},
		{twoFolders, "example", "example", false},
		{twoFolders, "another", "another", false},
	}

	for _, test := range tests {
		f, err := getEnvironmentFolder(test.repo, test.checkFolder)
		if test.expectedError && err == nil {
			t.Errorf("getEnvironmentFolder(%+v, '%v') gave nil err, but we wanted an error.", test.repo, test.checkFolder)
		} else if !test.expectedError && err != nil {
			t.Errorf("getEnvironmentFolder(%+v, '%v') gave error '%v', but we wanted no error.", test.repo, test.checkFolder, err)
		}
		if f != test.expectedFolder {
			t.Errorf("getEnvironmentFolder(%+v, '%v') gave folder '%v', but we wanted '%v'.", test.repo, test.checkFolder, f, test.expectedFolder)
		}
	}
}
