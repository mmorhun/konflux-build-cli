package integration_tests

import (
	"fmt"
	"os"
	"path"

	. "github.com/onsi/gomega"
)

const (
	KonfluxCli             = "konflux-task-cli"
	ResultsPathInContainer = "/tmp/"
)

// Set it true if you need to debug cli in container.
// Note, if set to true, cli in container will wait until debugger connects.
var Debug bool = false

func fileExists(filepath string) bool {
	stat, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

func konfluxCliCompiled() bool {
	return fileExists(path.Join("../", KonfluxCli))
}

func ExpectKonfluxCliCompiled() {
	Expect(konfluxCliCompiled()).To(BeTrue(), "CLI is not compiled. Compile it before running the test.")
}

func getDlvPath() (string, error) {
	goPath, isSet := os.LookupEnv("GOPATH")
	if !isSet {
		goPath = "~/go"
	}
	dlvPath := path.Join(goPath, "bin", "dlv")
	if !fileExists(dlvPath) {
		return "", fmt.Errorf("dlv is not found")
	}
	return dlvPath, nil
}

// CreateTempDir creates a directory in OS temp dir with given prefix
// and returns full path to the creted directory.
func CreateTempDir(prefix string) string {
	tmpDir, err := os.MkdirTemp("", prefix)
	Expect(err).ToNot(HaveOccurred())
	err = os.Chmod(tmpDir, 0777)
	Expect(err).ToNot(HaveOccurred())
	return tmpDir
}

func SaveToTempFile(data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "tmp-*")
	if err != nil {
		return "", err
	}
	if _, err := tmpFile.Write(data); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}
