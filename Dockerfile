# 使用官方 Golang Alpine 轻量镜像作为构建阶段基础镜像，标签为 builder
FROM golang:alpine AS builder

# 禁用 CGO，编译为纯静态链接的可执行文件（可跨平台运行）
ENV CGO_ENABLED 0

# 设置 Go 模块代理，加快依赖下载速度（可选：直连）
ENV GOPROXY https://goproxy.cn,direct

# 替换 Alpine 的默认软件源为阿里云，提高 apk 安装速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置工作目录为 /build（构建目录）
WORKDIR /build

# 将当前目录（宿主机代码）添加到容器内的 /build 目录
ADD . .

# 执行 Go 构建命令，生成 main 可执行文件
RUN go build -o main


# 构建阶段结束后，使用纯净的 Alpine 镜像作为最终镜像（瘦镜像）
FROM alpine

# 设置应用工作目录为 /app
WORKDIR /app

# 从 builder 阶段复制编译好的 main 文件到当前镜像的 /app 目录
COPY --from=builder /build/main /app

# 同样替换软件源为阿里云，提高 apk 安装速度
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装时区数据包，以支持设置时区（如 Asia/Shanghai）
RUN apk add tzdata

# 启动容器时默认执行 ./main 程序
CMD ["./main"]