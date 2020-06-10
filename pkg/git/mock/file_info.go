package mock

import (
	"os"
	"path/filepath"
	"time"
)

type mockFileInfo struct {
	name string
}

func newFileInfo(name string) mockFileInfo {
	return mockFileInfo{name}
}

func (f mockFileInfo) Name() string {
	return filepath.Base(f.name)
}

func (f mockFileInfo) Size() int64 {
	return 8
}

func (f mockFileInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (f mockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f mockFileInfo) IsDir() bool {
	return true
}

func (f mockFileInfo) Sys() interface{} {
	return nil
}
