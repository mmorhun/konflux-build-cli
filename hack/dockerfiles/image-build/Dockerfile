FROM konflux-task-cli:local AS source

FROM quay.io/konflux-ci/buildah-task:latest@sha256:cb58912cc9aecdb4c64e353ac44d0586574e89ba6cec2f2b191b4eeb98c6f81e
COPY --from=source /konflux-task-cli /usr/bin/konflux-task-cli
