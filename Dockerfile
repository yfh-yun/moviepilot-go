# 多阶段构建
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包（包含构建 CGO 所需工具）
RUN apk add --no-cache git ca-certificates tzdata build-base

# 设置Go代理（使用国内镜像源）
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
ENV GO111MODULE=on
ENV GOSUMDB=off

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码（包含预生成的 docs）
COPY . .

# 构建应用（启用 CGO，支持 go-sqlite3）
RUN go build -o main ./cmd/server

# 最终镜像
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S moviepilot && \
    adduser -u 1001 -S moviepilot -G moviepilot

WORKDIR /app

# 从builder阶段复制可执行文件
COPY --from=builder /app/main .

# 创建必要的目录
RUN mkdir -p /app/configs /app/logs /app/data && \
    chown -R moviepilot:moviepilot /app

# 切换到非root用户
USER moviepilot

# 暴露端口
EXPOSE 3001

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# 启动应用
CMD ["./main"]