package git

import (
	"io"
	"os"
)

// Destination is implemented by values that can have files written and stored
// permanently.
type Destination interface {
	CopyFile(src, dst string) error
	WriteFile(src io.Reader, dst string) error
}

// Source is implemented by values that can provide a list of files for reading
// from.
type Source interface {
	// Walk walks the repository tree, calling the callback with the prefix and
	// filename.
	Walk(path string, cb func(prefix, name string) error) error
	GetName() string
}

type Repo interface {
	Destination
	Source
	Clone() error
	Checkout(branch string) error
	CheckoutAndCreate(branch string) error
	DirectoriesUnderPath(path string) ([]os.FileInfo, error)
	GetUniqueEnvironmentFolder() (os.FileInfo, error)
	GetCommitID() string
	StageFiles(filenames ...string) error
	Commit(msg string, author *Author) error
	Push(branch string) error
	DeleteCache() error
}
