#!/bin/bash

IMAGE_REPOSITORY_PREFIX=quay.io/mmorhun-org

SHOULD_PUSH="true"

# Always run from the repository root
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

# Build the CLI and use the image as the CLI binary source
docker build -t konflux-task-cli:local -f hack/dockerfiles/Dockerfile .

git_clone_image="${IMAGE_REPOSITORY_PREFIX}/git-clone"
docker build -t "$git_clone_image" - < hack/dockerfiles/git-clone/Dockerfile
if [[ "$SHOULD_PUSH" == "true" ]]; then
    docker push "$git_clone_image"
fi

image_build_image="${IMAGE_REPOSITORY_PREFIX}/image-build"
docker build -t "$image_build_image" - < hack/dockerfiles/image-build/Dockerfile
if [[ "$SHOULD_PUSH" == "true" ]]; then
    docker push "$image_build_image"
fi

apply_tags_image="${IMAGE_REPOSITORY_PREFIX}/apply-tags"
docker build -t "$apply_tags_image" - < hack/dockerfiles/apply-tags/Dockerfile
if [[ "$SHOULD_PUSH" == "true" ]]; then
    docker push "$apply_tags_image"
fi
