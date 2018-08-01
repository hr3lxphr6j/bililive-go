BIN_PATH=./bin
BIN_NAME=bililive-go
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date '+%Y-%m-%d_%H:%M:%S')
GIT_HASH=$(shell git rev-parse HEAD)
CONSTS_PATH=github.com/hr3lxphr6j/bililive-go/src/consts
LD_FLAGS=-ldflags '-X ${CONSTS_PATH}.BuildTime=${BUILD_TIME} -X ${CONSTS_PATH}.AppVersion=${VERSION} -X ${CONSTS_PATH}.GitHash=${GIT_HASH}'

.PHONY: all
all: mkdir
	go build -o '${BIN_PATH}/${BIN_NAME}' ${LD_FLAGS} ./

.PHONY: mkdir
mkdir:
	@mkdir -p ${BIN_PATH}

.PHONY: release-all-platform-package
release-all-platform-package: release-all-platform
	for file in `ls ${BIN_PATH}`; \
	do \
		7z a ${BIN_PATH}/$${file%.*}.7z ${BIN_PATH}/$${file} ./config.yml; \
		rm ${BIN_PATH}/$${file}; \
	done

.PHONY: release-all-platform
release-all-platform: release-darwin-amd64 \
	release-freebsd-386 \
	release-freebsd-amd64 \
	release-linux-386 \
	release-linux-amd64 \
	release-linux-arm \
	release-linux-arm64 \
	release-linux-mips \
	release-linux-mips64 \
	release-linux-mipsle \
	release-linux-mips64le \
	release-windows-386 \
	release-windows-amd64

.PHONY: release-darwin-amd64
release-darwin-amd64: mkdir
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-darwin-amd64

.PHONY: release-freebsd-386
release-freebsd-386: mkdir
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-freebsd-386

.PHONY: release-freebsd-amd64
release-freebsd-amd64: mkdir
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-freebsd-amd64

.PHONY: release-linux-386
release-linux-386: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-386

.PHONY: release-linux-amd64
release-linux-amd64: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-amd64

.PHONY: release-linux-arm
release-linux-arm: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-arm

.PHONY: release-linux-arm64
release-linux-arm64: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-arm64

.PHONY: release-linux-mips
release-linux-mips: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=mips go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-mips

.PHONY: release-linux-mips64
release-linux-mips64: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-mips64

.PHONY: release-linux-mipsle
release-linux-mipsle: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=mipsle go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-mipsle

.PHONY: release-linux-mips64le
release-linux-mips64le: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64le go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-mips64le

.PHONY: release-windows-386
release-windows-386: mkdir
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-windows-386.exe

.PHONY: release-windows-amd64
release-windows-amd64: mkdir
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-windows-amd64.exe

.PHONY: release-docker
release-docker: release-linux-amd64
	docker build -t 'bililive-go:${VERSION}' .

.PHONY: test
test:
	@go test --cover ./src/api

.PHONY: clean
clean:
	@rm -rf ${BIN_PATH}
	@echo "All clean"