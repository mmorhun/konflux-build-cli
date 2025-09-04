FROM konflux-task-cli:local AS source

FROM quay.io/konflux-ci/git-clone@sha256:4e53ebd9242f05ca55bfc8d58b3363d8b9d9bc3ab439d9ab76cdbdf5b1fd42d9
COPY --from=source /konflux-task-cli /usr/bin/konflux-task-cli
