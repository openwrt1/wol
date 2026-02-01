# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN go build -o wol .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/wol .

EXPOSE 7777

ENTRYPOINT ["/app/wol"]