FROM registry.access.redhat.com/ubi9/go-toolset:1.24.6-1756993846 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0
RUN go build -o konflux-task-cli main.go

FROM scratch
COPY --from=builder /src/konflux-task-cli /konflux-task-cli
