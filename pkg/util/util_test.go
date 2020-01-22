package util

import "testing"

func TestExtractUserAndRepo(t *testing.T) {
	urlTests := []struct {
		url  string
		user string
		repo string
	}{
		{"https://github.com/user/testing.git", "user", "testing"},
	}

	for _, tt := range urlTests {
		user, repo, err := ExtractUserAndRepo(tt.url)
		if err != nil {
			t.Errorf("ExtractUserAndRepo(%v) got an error: %s", tt.url, err)
		}
		if tt.user != user {
			t.Errorf("ExtractUserAndRepo(%v) got user %s, want %s", tt.url, user, tt.user)
		}
		if tt.repo != repo {
			t.Errorf("ExtractUserAndRepo(%v) got repo %s, want %s", tt.url, repo, tt.repo)
		}
	}

}
