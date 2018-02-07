#!/usr/bin/env bash

out_path="./bins"
bin_name="bililive-go"

rm -rf ${out_path}
mkdir ${out_path}

env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_darwin_amd64
env CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_freebsd_386
env CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_freebsd_amd64
env CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_386
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_amd64
env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_arm
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_arm64
env CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_mips64
env CGO_ENABLED=0 GOOS=linux GOARCH=mips64le go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_mips64le
env CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_mips
env CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_linux_mipsle
env CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_windows_386.exe
env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ${out_path}/${bin_name}_windows_amd64.exe

for file in `ls ${out_path}`
do
 7z a ${out_path}/${file%.*}.7z ${out_path}/${file} ./config.yml
 rm ${out_path}/${file}
done