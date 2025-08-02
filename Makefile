PROJECT_NAME=cloudpan189-share
MODULE_NAME=github.com/xxcheng123/cloudpan189-share
VAR_COMMIT ?= $(shell git rev-parse HEAD)
VAR_BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VAR_GIT_SUMMARY ?= $(shell git describe --tags --dirty --always )
VAR_GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)"  -o output/subscribe ./cmd/main.go


.PHONY: clear
clear:
	golangci-lint cache clean

.PHONY: docker-build
docker-build:
	docker build \
		--build-arg MODULE_NAME=$(MODULE_NAME) \
		--build-arg VAR_COMMIT=$(VAR_COMMIT) \
		--build-arg VAR_BUILD_DATE=$(VAR_BUILD_DATE) \
		--build-arg VAR_GIT_SUMMARY=$(VAR_GIT_SUMMARY) \
		--build-arg VAR_GIT_BRANCH=$(VAR_GIT_BRANCH) \
		-t cloudpan189-share:latest .

.PHONY: docker-run
docker-run:
	docker run -d -p 12395:12395 --name cloudpan189-share cloudpan189-share:latest

.PHONY: clean-docker
clean-docker:
	docker rm -f cloudpan189-share || true
	docker rmi cloudpan189-share:latest || true