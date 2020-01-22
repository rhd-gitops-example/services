package cache

import "testing"

func TestEncodeURL(t *testing.T) {
	urlTests := []struct {
		url  string
		want string
	}{
		{url: "https://github.com/testing/testing.git", want: "https%3A%2F%2Fgithub.com%2Ftesting%2Ftesting.git"},
	}

	for _, tt := range urlTests {
		if e := encodeURL(tt.url); e != tt.want {
			t.Errorf("encodeURL(%v) got %v, want %v", tt.url, e, tt.want)
		}
	}

}

func TestDirExists(t *testing.T) {
	t.Skip()
}
