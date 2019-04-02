BIN_PATH=./bin
BIN_NAME=bililive-go
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date '+%Y-%m-%d_%H:%M:%S')
GIT_HASH=$(shell git rev-parse HEAD)
CONSTS_PATH=github.com/hr3lxphr6j/bililive-go/src/consts
LD_FLAGS=-ldflags '-s -w -X ${CONSTS_PATH}.BuildTime=${BUILD_TIME} -X ${CONSTS_PATH}.AppVersion=${VERSION} -X ${CONSTS_PATH}.GitHash=${GIT_HASH}'
UPX=1

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
	release-linux-386 \
	release-linux-amd64 \
	release-linux-arm \
	release-linux-arm64 \
	release-windows-386 \
	release-windows-amd64

.PHONY: release-darwin-amd64
release-darwin-amd64: mkdir
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-darwin-amd64
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-darwin-amd64; \
	fi

.PHONY: release-linux-386
release-linux-386: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-386
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-linux-386; \
	fi

.PHONY: release-linux-amd64
release-linux-amd64: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-amd64
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-linux-amd64; \
	fi

.PHONY: release-linux-arm
release-linux-arm: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-arm
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-linux-arm; \
	fi

.PHONY: release-linux-arm64
release-linux-arm64: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-linux-arm64
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-linux-arm64; \
	fi

.PHONY: release-windows-386
release-windows-386: mkdir
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-windows-386.exe
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-windows-386.exe; \
	fi

.PHONY: release-windows-amd64
release-windows-amd64: mkdir
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LD_FLAGS} -o ${BIN_PATH}/${BIN_NAME}-windows-amd64.exe
	@if [ ${UPX} -eq 1 ]; \
    then \
		upx --no-progress ${BIN_PATH}/${BIN_NAME}-windows-amd64.exe; \
	fi

.PHONY: release-docker
release-docker: release-linux-amd64
	docker build -t 'chigusa/bililive-go' .

.PHONY: test
test:
	@go test --cover ./src/api

.PHONY: clean
clean:
	@rm -rf ${BIN_PATH}
	@echo "All clean"