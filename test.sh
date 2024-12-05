ARCH=$(uname -m)

# 设置下载的URL
if [ "$ARCH" = "x86_64" ]; then
    URL="https://nodejs.org/dist/latest/node-v14.17.3-linux-x64.tar.xz"
elif [ "$ARCH" = "aarch64" ]; then
    URL="https://nodejs.org/dist/latest/node-v14.17.3-linux-arm64.tar.xz"
else
    echo "不支持的架构: $ARCH"
    exit 1
fi

echo $URL