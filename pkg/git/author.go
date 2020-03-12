package git

// Author represents the details needed to commit and push to Git, including the
// auth-token to use when pushing upstream.
type Author struct {
	Name  string
	Email string
	Token string
}
