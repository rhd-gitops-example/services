package git

import (
	"bytes"
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

type Repository struct {
	LocalPath string
	RepoURL   string
	repoName  string
	debug     bool
	logger    func(fmt string, v ...interface{})
}

// NewRepository creates and returns a local cache of an upstream repository.
//
// The repoURL should be of the form https://github.com/myorg/myrepo.
// The path should be a local directory where the contents are cloned to.
func NewRepository(repoURL, localPath string, debug bool) (*Repository, error) {
	name, err := repoName(repoURL)
	if err != nil {
		return nil, err
	}
	return &Repository{LocalPath: localPath, RepoURL: repoURL, repoName: name, logger: log.Printf, debug: debug}, nil
}

func (g *Repository) repoPath(extras ...string) string {
	fullPath := append([]string{g.LocalPath, g.repoName}, extras...)
	return path.Join(fullPath...)
}

func (r *Repository) Clone() error {
	err := os.MkdirAll(r.LocalPath, 0755)
	if err != nil {
		return fmt.Errorf("error creating the cache dir %s: %w", r.LocalPath, err)
	}
	out, err := r.execGit(r.LocalPath, nil, "clone", r.RepoURL)
	log.Printf("%s\n", out)
	return err
}

func (r *Repository) Checkout(branch string) error {
	_, err := r.execGit(r.repoPath(), nil, "checkout", branch)
	return err
}

func (r *Repository) CheckoutAndCreate(branch string) error {
	_, err := r.execGit(r.repoPath(), nil, "checkout", "-b", branch)
	return err
}

func (r *Repository) GetName() string {
	return r.repoName
}

func (r *Repository) GetCommitID() string {
	commitID, _ := r.execGit(r.repoPath(), nil, "rev-parse", "--short", "HEAD")
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
	_, err := r.execGit(r.repoPath(), nil, append([]string{"add"}, filenames...)...)
	return err
}

func (r *Repository) Commit(msg string, author *Author) error {
	args := []string{"commit", "-m", msg}
	_, err := r.execGit(r.repoPath(), envFromAuthor(author), args...)
	return err
}

func (r *Repository) Push(branchName string) error {
	args := []string{"push", "origin", branchName}
	_, err := r.execGit(r.repoPath(), nil, args...)
	return err
}

func (r *Repository) execGit(workingDir string, env []string, args ...string) ([]byte, error) {
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
	if err != nil && r.debug {
		r.logger("DEBUG: %s\n", out)
	}
	return out, err
}

// TODO: this probably needs specialisation for GitLab URLs.
// TODO: should we process "git@github.com" urls?, this would require SSH keys.
func repoName(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL '%s': %w", u, err)
	}
	parts := strings.Split(parsed.Path, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("could not identify repo: %s", u)
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
