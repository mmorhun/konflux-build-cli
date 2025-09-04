#!/bin/sh

export WORKDIR=$(mktemp -d)
cd "$WORKDIR"

# Params
export GIT_REPO_URL="https://github.com/devfile-samples/devfile-sample-go-basic"

# Results
RESULTS_DIR="$WORKDIR/results"
mkdir -p "$RESULTS_DIR"
export RESULT_URL="${PRESULTS_DIR}/RESULT_URL"
export RESULT_SOURCE_DIR="${PRESULTS_DIR}/RESULT_SOURCE_DIR"
export RESULT_COMMIT="${RESULTS_DIR}/RESULT_COMMIT"
export RESULT_SHORT_COMMIT="${RESULTS_DIR}/RESULT_SHORT_COMMIT"

# konflux-task-cli gitclone
