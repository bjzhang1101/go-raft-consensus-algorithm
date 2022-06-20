.PHONY: build
build:
	@go build -o raft main.go

.PHONY: build-linux
build-linux:
	@GOOS=linux GOARCH=amd64 go build -o raft-linux main.go

# Add dependencies if not installed.
GOIMPORTS_EXISTS := $(shell command -v goimports 2> /dev/null)
.PHONY: dependencies
dependencies:
	@if [ -z "$(GOIMPORTS_EXISTS)" ] ; then go install golang.org/x/tools/cmd/goimports@latest ; fi

.PHONY: fmt
fmt: dependencies
	@goimports -local=airbnb.com -w .

.PHONY: tidy
tidy:
	@go mod tidy
