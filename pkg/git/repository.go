package git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var _ Repo = (*Repository)(nil)

type loggerFunc func(fmt string, v ...interface{})

type Repository struct {
	LocalPath string
	RepoURL   string
	repoName  string
	verifyTLS bool
	debug     bool
	logger    loggerFunc
}

// NewRepository creates and returns a local cache of an upstream repository.
//
// The repoURL should be of the form https://github.com/myorg/myrepo.
// The path should be a local directory where the contents are cloned to.
func NewRemoteRepository(repoURL, cacheDir string, verifyTLS bool, debug bool) (*Repository, error) {
	localPath, err := cloneRepo(cacheDir, repoURL, verifyTLS, debug, log.Printf)
	if err != nil {
		return nil, err
	}
	return &Repository{LocalPath: localPath, RepoURL: repoURL, logger: log.Printf, debug: debug, verifyTLS: verifyTLS}, nil
}

func (g *Repository) repoPath(extras ...string) string {
	fullPath := append([]string{g.LocalPath}, extras...)
	return path.Join(fullPath...)
}

func (r *Repository) Checkout(branch string) error {
	_, err := execGit(r.repoPath(), nil, r.verifyTLS, r.debug, r.logger, "checkout", branch)
	return err
}

func (r *Repository) CheckoutAndCreate(branch string) error {
	_, err := execGit(r.repoPath(), nil, r.verifyTLS, r.debug, r.logger, "checkout", "-b", branch)
	return err
}

func (r *Repository) GetName() string {
	return ""
}

func (r *Repository) CommitID() string {
	commitID, _ := execGit(r.repoPath(), nil, r.verifyTLS, r.debug, r.logger, "rev-parse", "--short", "HEAD")
	return string(commitID)
}

func (r *Repository) Walk(base string, cb func(prefix, name string) error) error {
	repoBase := r.repoPath(base)
	prefix := filepath.Dir(repoBase) + "/"
	return filepath.Walk(repoBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		return cb(prefix, strings.TrimPrefix(path, prefix))
	})
}

func (r *Repository) WriteFile(src io.Reader, dst string) error {
	filename := r.repoPath(dst)
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", filename, err)
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func (r *Repository) CopyFile(src, dst string) error {
	outputPath := r.repoPath(dst)
	err := os.MkdirAll(path.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed trying to create directory for file copy %s: %w", path.Dir(outputPath), err)
	}
	return fileCopy(src, outputPath)
}

func (r *Repository) StageFiles(filenames ...string) error {
	_, err := execGit(r.repoPath(), nil, r.verifyTLS, r.debug, r.logger, append([]string{"add"}, filenames...)...)
	return err
}

func (r *Repository) Commit(msg string, author *Author) error {
	args := []string{"commit", "-m", msg}
	_, err := execGit(r.repoPath(), envFromAuthor(author), r.verifyTLS, r.debug, r.logger, args...)
	return err
}

func (r *Repository) Push(branchName string) error {
	args := []string{"push", "origin", branchName}
	_, err := execGit(r.repoPath(), nil, r.verifyTLS, r.debug, r.logger, args...)
	return err
}

// TODO: this probably needs specialisation for GitLab URLs.
// TODO: should we process "git@github.com" urls?, this would require SSH keys.
func repoName(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		// Don't surface the URL as it could contain a token
		return "", errors.New("failed to parse the URL when determining repository name")
	}
	parts := strings.Split(parsed.Path, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("could not identify repository name: %s", u)
	}
	return strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
}

// To  clone with a username/password
// git clone https://username:password@github.com/username/repository.git

func envFromAuthor(a *Author) []string {
	sf := func(k, v string) string {
		return fmt.Sprintf("%s=%s", k, v)
	}
	return []string{
		sf("GIT_AUTHOR_NAME", a.Name),
		sf("GIT_AUTHOR_EMAIL", a.Email)}
}

// DeleteCache removes the local clones from the promotion cache.
func (r *Repository) DeleteCache() error {
	err := os.RemoveAll(r.LocalPath)
	if err != nil {
		return fmt.Errorf("failed deleting `%s` : %w", r.LocalPath, err)
	}
	return nil
}

func cloneRepo(localPath, repoURL string, verifyTLS, debug bool, logger loggerFunc) (string, error) {
	name, err := repoName(repoURL)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(localPath, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating the cache dir %s: %w", localPath, err)
	}
	// Intentionally omit output as it can contain an access token.
	_, err = execGit(localPath, nil, verifyTLS, debug, logger, "clone", repoURL)
	return filepath.Join(localPath, name), err
}

func execGit(workingDir string, env []string, verifyTLS, debug bool, logger loggerFunc, args ...string) ([]byte, error) {
	if verifyTLS {
		env = append(env, "GIT_SSL_NO_VERIFY=true")
	}
	cmd := exec.Command("git", args...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	cmd.Dir = workingDir
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()
	out := b.Bytes()
	// TODO: more sophisticated logging.
	if err != nil && debug {
		logger("DEBUG: %s\n", out)
	}
	return out, err
}
