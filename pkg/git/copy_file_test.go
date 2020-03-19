package git

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var testContent = []byte(`test content`)

// TODO: fileCopy does a number of things, and catches a lot of error cases.
// Testing the error cases is non-trivial, getting it to fail in some of the
// more obscure cases is tricky.
func TestFileCopy(t *testing.T) {
	tempDir, cleanup := makeTempDir(t)
	defer cleanup()
	srcFile := path.Join(tempDir, "myfile.txt")
	dstFile := path.Join(tempDir, "newfile.txt")
	writeFile(t, srcFile, 0644)

	err := fileCopy(srcFile, dstFile)
	assertNoError(t, err)

	assertFileMode(t, dstFile, 0644)
	assertFileContents(t, dstFile, testContent)
}

func writeFile(t *testing.T, filename string, perms os.FileMode) {
	t.Helper()
	if err := ioutil.WriteFile(filename, testContent, perms); err != nil {
		t.Fatalf("failed to write file to %s: %s", filename, err)
	}
}

func assertFileMode(t *testing.T, filename string, perms os.FileMode) {
	t.Helper()
	mode, err := fileMode(filename)
	if err != nil {
		t.Fatalf("failed get file permissions for %s: %s", filename, err)
	}
	if mode != perms {
		t.Fatalf("incorrect perms on file, got %v, want %v", mode, perms)
	}
}

func assertFileContents(t *testing.T, filename string, want []byte) {
	contents, err := ioutil.ReadFile(filename)
	assertNoError(t, err)
	if !reflect.DeepEqual(contents, want) {
		t.Fatalf("contents read from file did not match, got %#v, want %#v", contents, want)
	}
}
