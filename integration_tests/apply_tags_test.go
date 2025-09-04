package integration_tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

const ApplyTagsImage = "registry.access.redhat.com/ubi9/skopeo:9.6-1754871306@sha256:e59e2cb3fd8d7613798738fb06aad5aab61f32c18aed595df16a46a8e078dfa6"

type ApplyTagsParams struct {
	ImageRepoUrl string
	ImageDigest  string
	Tags         []string
}

func RunApplyTags(applyTagsParams ApplyTagsParams, login, password string) error {
	var err error

	container := NewTestContainer("apply-tags", ApplyTagsImage, true)

	// Params
	container.AddEnv("IMAGE_URL", applyTagsParams.ImageRepoUrl)
	container.AddEnv("IMAGE_DIGEST", applyTagsParams.ImageDigest)
	// Tags are passed via CLI to test it
	container.AddEnv("VERBOSE", "true")

	if Debug {
		container.AddPort("2345", "2345")
	}
	err = container.Start()
	Expect(err).ToNot(HaveOccurred())
	defer container.Delete()

	err = container.CopyFileIntoContainer("../"+KonfluxCli, "/usr/bin/")
	Expect(err).ToNot(HaveOccurred())

	// err = container.InjectDockerAuth(applyTagsParams.ImageRepoUrl, login, password)
	err = container.InjectDockerAuth("quay.io", login, password)
	Expect(err).ToNot(HaveOccurred())

	args := []string{"image", "apply-tags"}
	if len(applyTagsParams.Tags) > 0 {
		args = append(args, "--tags")
		args = append(args, applyTagsParams.Tags...)
	}

	if Debug {
		err = container.DebugCli(args...)
	} else {
		err = container.ExecuteAndWait(KonfluxCli, args...)
	}
	Expect(err).ToNot(HaveOccurred())

	return nil
}

func TestApplyTags(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		fmt.Printf("Test Failure: %s\n", message)
		t.FailNow() // Terminate the test immediately
	})
	ExpectKonfluxCliCompiled()

	const imageRepoUrl = "quay.io/mmorhun-org/test-build-repo"
	const imageDigest = "sha256:0acffdabda074fb9ab4b9fc38c049db903d2199b0a8be64ee0f1ca5a4fb74667"
	newTag := time.Now().Format("2006-01-02_15-04-05")
	applyTagsParams := ApplyTagsParams{
		ImageRepoUrl: imageRepoUrl,
		ImageDigest:  imageDigest,
		Tags:         []string{newTag, "latest"},
	}

	login := os.Getenv("QUAY_ROBOT_NAME")
	password := os.Getenv("QUAY_ROBOT_TOKEN")
	Expect(login).ToNot(BeEmpty(), "QUAY_ROBOT_NAME is not set")
	Expect(password).ToNot(BeEmpty(), "QUAY_ROBOT_TOKEN is not set")

	err := RunApplyTags(applyTagsParams, login, password)
	Expect(err).ToNot(HaveOccurred())

	imageByTagFromParams := imageRepoUrl + ":" + newTag
	imageByTagFromParamsExists, err := CheckTagExistance(imageByTagFromParams, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageByTagFromParamsExists).To(BeTrue())

	imageByLatestTagFromParams := imageRepoUrl + ":latest"
	imageByLatestTagFromParamsExists, err := CheckTagExistance(imageByLatestTagFromParams, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageByLatestTagFromParamsExists).To(BeTrue())

	imageByLabelInImage := imageRepoUrl + ":label-tag"
	imageByLabelInImageExists, err := CheckTagExistance(imageByLabelInImage, login, password)
	Expect(err).ToNot(HaveOccurred())
	Expect(imageByLabelInImageExists).To(BeTrue())
}

// CheckTagExistance quaries Quay API to check the tag existance.
// image has format: quay.io/namespace/repo:tag
func CheckTagExistance(image, username, password string) (bool, error) {
	parts := strings.Split(image, ":")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid image format, expected quay.io/namespace/repo:tag")
	}
	repo := parts[0]
	tag := parts[1]

	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 3 {
		return false, fmt.Errorf("invalid image format, expected quay.io/namespace/repo")
	}
	namespace := repoParts[1]
	repository := repoParts[2]

	url := fmt.Sprintf("https://quay.io/api/v1/repository/%s/%s/tag/?specificTag=%s", namespace, repository, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	// {
	//   "tags": [
	//     {
	//       "name": "tag-name",
	//       "reversion": false,
	//       "start_ts": 1756740181,
	//       "manifest_digest": "sha256:33735bd63cf84d7e388d9f6d297d348c523c044410f553bd878c6d7829612735",
	//       "is_manifest_list": false,
	//       "size": 3623807,
	//       "last_modified": "Mon, 01 Sep 2025 15:23:01 -0000"
	//     }
	//   ]
	// }
	type Tag struct {
		Name string `json:"name"`
	}
	type Response struct {
		Tags []Tag `json:"tags"`
	}
	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, err
	}

	for _, t := range result.Tags {
		if t.Name == tag {
			return true, nil
		}
	}
	return false, nil
}
