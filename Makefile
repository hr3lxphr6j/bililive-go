UPX_ENABLE ?= 0
PLATFORM ?= $(shell go env GOHOSTOS)
ARCH ?= $(shell go env GOHOSTARCH)
local_go_version := $(shell go version | cut -d' ' -f3 | sed -e 's/go//g')
tags := "release"

build: $(notdir $(abspath $(wildcard src/cmd/*/)))
.PHONY: build

$(notdir $(abspath $(wildcard src/cmd/*/))):
	@echo "building $@ (Platform: $(PLATFORM), Arch: $(ARCH), GoVersion: $(local_go_version))"
	@GOOS=$(PLATFORM) \
		GOARCH=$(ARCH) \
		CGO_ENABLED=0 \
		UPX_ENABLE=$(UPX_ENABLE) \
		TAGS=$(tags) \
		GCFLAGS=$(gcflags) \
		./src/hack/build.sh $@

.PHONY: dev
dev: tags := "dev"
dev: gcflags := "all=-N -l"
dev: build

.PHONY: release
release: build-web generate
	@./src/hack/release.sh

.PHONY: release-docker
release-docker:
	@./src/hack/release-docker.sh

.PHONY: test
test:
	@go test -tags release --cover -coverprofile=coverage.txt ./src/...

.PHONY: clean
clean:
	@rm -rf bin ./src/webapp/build
	@echo "All clean"

.PHONY: generate
generate:
	go generate ./...

.PHONY: build-web
build-web:
	cd ./src/webapp && yarn install && yarn build && cd ../../

.PHONY: run
run:
	foreman start || exit 0
