.PHONY: build
build:
	@go build -o raft *.go

.PHONY: build-linux
build-linux:
	@GOOS=linux GOARCH=amd64 go build -o raft-linux main.go

# Add dependencies if not installed.
GOIMPORTS_EXISTS := $(shell command -v goimports 2> /dev/null)
.PHONY: dependencies
dependencies:
	@if [ -z "$(GOIMPORTS_EXISTS)" ] ; then go install golang.org/x/tools/cmd/goimports@latest ; fi

.PHONY: docker-build
docker-build:
	@eval $(minikube docker-env)
	@docker build --rm -t raft:$(tag) .
	# Prune intermediate builder image.
	@docker image prune -f --filter label=stage=builder

.PHONY: fmt
fmt: dependencies
	@goimports -local=github.com/bjzhang1101/raft -w .

.PHONY: kube-apply
	@kubectl apply -f k8s-manifests

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: proto-gen
proto-gen:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/protobuf/*.proto
