package integration_tests

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

const ImageBuildImage = "quay.io/konflux-ci/buildah-task:latest@sha256:cb58912cc9aecdb4c64e353ac44d0586574e89ba6cec2f2b191b4eeb98c6f81e"

type ImageBuildParams struct {
	Image       string
	SourceDir   string
	Dockerfile  string
	labels      []string
	annotations []string
}

type ImageBuildResults struct {
	Url    string
	Digest string
}

func RunImageBuild(imageBuildParams ImageBuildParams, login, password string, volumeHostPath string) (ImageBuildResults, error) {
	var err error

	container := NewTestContainer("image-build", ImageBuildImage, true)
	container.privileged = true

	// Params
	container.AddEnv("IMAGE", imageBuildParams.Image)
	container.AddEnv("SOURCE_DIR", imageBuildParams.SourceDir)
	container.AddEnv("DOCKERFILE", imageBuildParams.Dockerfile)
	container.AddEnv("LABELS", strings.Join(imageBuildParams.labels, " "))
	container.AddEnv("ANNOTATIONS", strings.Join(imageBuildParams.annotations, " "))
	container.AddEnv("VERBOSE", "true")
	// Results
	container.AddTaskResult("RESULT_IMAGE_URL")
	container.AddTaskResult("RESULT_IMAGE_DIGEST")

	if volumeHostPath != "" {
		container.AddVolume(volumeHostPath, "/pvc")
		container.SetWorkdir("/pvc")
	}

	if Debug {
		container.AddPort("2345", "2345")
	}
	err = container.Start()
	Expect(err).ToNot(HaveOccurred())
	defer container.Delete()

	err = container.CopyFileIntoContainer("../"+KonfluxCli, "/usr/bin/")
	Expect(err).ToNot(HaveOccurred())

	err = container.InjectDockerAuth("quay.io", login, password)
	Expect(err).ToNot(HaveOccurred())

	if Debug {
		err = container.DebugCli("image", "build")
	} else {
		err = container.ExecuteAndWait(KonfluxCli, "image", "build")
	}
	Expect(err).ToNot(HaveOccurred())

	imageUrl, err := container.GetTaskResultValue("RESULT_IMAGE_URL")
	Expect(err).ToNot(HaveOccurred())
	digest, err := container.GetTaskResultValue("RESULT_IMAGE_DIGEST")
	Expect(err).ToNot(HaveOccurred())

	return ImageBuildResults{
		Url:    imageUrl,
		Digest: digest,
	}, nil
}

func TestImageBuild(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		fmt.Printf("Test Failure: %s\n", message)
		t.FailNow() // Terminate the test immediately
	})
	ExpectKonfluxCliCompiled()

	login := os.Getenv("QUAY_ROBOT_NAME")
	password := os.Getenv("QUAY_ROBOT_TOKEN")
	Expect(login).ToNot(BeEmpty(), "QUAY_ROBOT_NAME is not set")
	Expect(password).ToNot(BeEmpty(), "QUAY_ROBOT_TOKEN is not set")

	volumeDir := CreateTempDir("build-pvc-*")
	defer os.RemoveAll(volumeDir)

	err := createProjectFiles(volumeDir)
	Expect(err).ToNot(HaveOccurred())

	const imageRepoUrl = "quay.io/mmorhun-org/test-build-repo"
	tag := "build-" + strconv.FormatInt(time.Now().Unix(), 10)
	imageBuildParams := ImageBuildParams{
		Image:       imageRepoUrl + ":" + tag,
		SourceDir:   ".",
		annotations: []string{"a1=v1", "a2=v2"},
	}

	imageBuildResults, err := RunImageBuild(imageBuildParams, login, password, volumeDir)
	Expect(err).ToNot(HaveOccurred())

	Expect(imageBuildResults.Url).To(HavePrefix("quay.io/mmorhun-org/test-build-repo:build-"))
	Expect(imageBuildResults.Digest).To(MatchRegexp(`^sha256:[0-9a-f]{64}$`))

	imageRef := imageBuildResults.Url + "@" + tag
	builtImageExists, err := CheckTagExistance(imageRef, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(builtImageExists).To(BeTrue())
}

func createProjectFiles(dir string) error {
	dockerfile := []byte(`
		FROM registry.access.redhat.com/ubi9/go-toolset:1.19.13-4.1697647145
		COPY . .
		RUN go mod download
		RUN go build ./main.go
		ENV PORT 8081
		EXPOSE 8081
		CMD [ "./main" ]
	`)
	goMain := []byte(`
		package main
		import "fmt"
		func main() {
			fmt.Println("Hello World!")
		}
	`)
	goMod := []byte(`module example.com/greetings
go 1.23
	`)

	fileMode := os.FileMode(0644)
	if err := os.WriteFile(path.Join(dir, "Dockerfile"), dockerfile, fileMode); err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(dir, "main.go"), goMain, fileMode); err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(dir, "go.mod"), goMod, fileMode); err != nil {
		return err
	}
	return nil
}
