FROM konflux-task-cli:local AS source

FROM registry.access.redhat.com/ubi9/skopeo:9.6-1754871306@sha256:e59e2cb3fd8d7613798738fb06aad5aab61f32c18aed595df16a46a8e078dfa6
COPY --from=source /konflux-task-cli /usr/bin/konflux-task-cli
