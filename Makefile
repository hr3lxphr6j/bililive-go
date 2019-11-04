UPX_ENABLE ?= 0
PLATFORM ?= $(shell go env GOHOSTOS)
ARCH ?= $(shell go env GOHOSTARCH)


.PHONY: $(notdir $(abspath $(wildcard src/cmd/*/)))
local_go_version := $(shell go version | cut -d' ' -f3 | sed -e 's/go//g')
$(notdir $(abspath $(wildcard src/cmd/*/))):
	@echo "building $@ (Platform: $(PLATFORM), Arch: $(ARCH), GoVersion: $(local_go_version))"
	@GOOS=$(PLATFORM) \
		GOARCH=$(ARCH) \
		CGO_ENABLED=0 \
		GOFLAGS=$(GOFLAGS) \
		UPX_ENABLE=$(UPX_ENABLE) \
		./src/hack/build.sh $@

.PHONY: release
release:
	@./src/hack/release.sh

.PHONY: release-docker
release-docker: clean
	@./src/hack/release-docker.sh

.PHONY: test
test:
	@go test --cover ./src/...

.PHONY: clean
clean:
	@rm -rf bin
	@echo "All clean"

.PHONY: generate
generate:
	go generate ./...