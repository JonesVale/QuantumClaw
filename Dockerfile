# QuantumClaw Open Source Edition

FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o quantumclaw ./cmd/server

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 创建日志目录
RUN mkdir -p /app/logs /app/config

# 从构建阶段复制二进制文件
COPY --from=builder /app/quantumclaw .

# 暴露端口
EXPOSE 3666

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3666/api/status || exit 1

# 运行
CMD ["./quantumclaw"]
