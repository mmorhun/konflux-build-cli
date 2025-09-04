
.PHONY: build
build:
	go build -o konflux-task-cli main.go

.PHONY: build-debug
build-debug:
	go build -gcflags "all=-N -l" -o konflux-task-cli main.go

.PHONY: test
test:
	go test ./pkg/...
