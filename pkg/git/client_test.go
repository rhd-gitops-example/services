package git

import (
	"testing"
)
func TestCreateClient(t *testing.T) {
	create := []struct {
		toUrl       string
		repoType  string
		wantHost  string
	}{
		{"https://github.com/myorg/myrepo.git", "github", "api.github.com"},
		{"https://gitlab.local.com/myorg/subdir/myrepo.git", "gitlab", "gitlab.local.com"},
		{"https://github.local.com/myorg/myrepo.git", "ghe", "github.local.com"},
	}

	for _, tt := range create {
		client := CreateClient("tokendata", tt.toUrl, tt.repoType, true)
		if tt.wantHost != client.BaseURL.Host {
			t.Errorf("CreateClient(%s) got host %s, want %s", tt.toUrl, client.BaseURL.Host, tt.wantHost)
			continue
		}
	}
}
