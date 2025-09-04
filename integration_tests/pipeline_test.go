package integration_tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestPipeline(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		fmt.Printf("Test Failure: %s\n", message)
		t.FailNow() // Terminate the test immediately
	})
	ExpectKonfluxCliCompiled()

	login := os.Getenv("QUAY_ROBOT_NAME")
	password := os.Getenv("QUAY_ROBOT_TOKEN")
	Expect(login).ToNot(BeEmpty(), "QUAY_ROBOT_NAME is not set")
	Expect(password).ToNot(BeEmpty(), "QUAY_ROBOT_TOKEN is not set")

	const repoUrl = "https://github.com/devfile-samples/devfile-sample-go-basic"
	const imageRepoUrl = "quay.io/mmorhun-org/test-build-repo"
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	tag := "build-goapp-" + timestamp
	additionalTagFromLabels := "additional-" + timestamp

	volumeDir := CreateTempDir("pipeline-pvc-*")
	defer os.RemoveAll(volumeDir)

	// Clone
	gitCloneParams := GitCloneParams{
		RepoUrl: repoUrl,
	}
	gitCloneResults, err := RunGitClone(gitCloneParams, volumeDir)
	Expect(err).ToNot(HaveOccurred())
	Expect(gitCloneResults.Url).To(Equal(repoUrl))
	Expect(gitCloneResults.SourceDir).To(Equal("devfile-sample-go-basic"))

	// Build
	imageBuildParams := ImageBuildParams{
		Image:      imageRepoUrl + ":" + tag,
		SourceDir:  gitCloneResults.SourceDir,
		Dockerfile: "docker/Dockerfile",
		labels:     []string{fmt.Sprintf("konflux.additional-tags=%s", additionalTagFromLabels)},
	}
	imageBuildResults, err := RunImageBuild(imageBuildParams, login, password, volumeDir)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageBuildResults.Url).To(HavePrefix("quay.io/mmorhun-org/test-build-repo:build-goapp-"))
	Expect(imageBuildResults.Digest).To(MatchRegexp(`^sha256:[0-9a-f]{64}$`))

	builtImageExists, err := CheckTagExistance(imageBuildResults.Url, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(builtImageExists).To(BeTrue(), fmt.Sprintf("built image %s does not exist in registry", imageBuildResults.Url))

	// Apply additional tags
	applyTagsParams := ApplyTagsParams{
		ImageRepoUrl: imageBuildResults.Url,
		ImageDigest:  imageBuildResults.Digest,
		Tags:         []string{timestamp, "latest"},
	}
	err = RunApplyTags(applyTagsParams, login, password)
	Expect(err).ToNot(HaveOccurred())

	imageByTagFromParamsRef := imageRepoUrl + ":" + timestamp
	imageByTagFromParamsExists, err := CheckTagExistance(imageByTagFromParamsRef, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageByTagFromParamsExists).To(BeTrue(), fmt.Sprintf("image %s does not exist", imageByTagFromParamsRef))

	imageByLatestTagFromParamsRef := imageRepoUrl + ":latest"
	imageByLatestTagFromParamsExists, err := CheckTagExistance(imageByLatestTagFromParamsRef, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageByLatestTagFromParamsExists).To(BeTrue(), fmt.Sprintf("image %s does not exist", imageByLatestTagFromParamsRef))

	imageByLabelInImageRef := imageRepoUrl + ":" + additionalTagFromLabels
	imageByLabelInImageExists, err := CheckTagExistance(imageByLabelInImageRef, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageByLabelInImageExists).To(BeTrue(), fmt.Sprintf("image %s does not exist", imageByLabelInImageRef))
}
