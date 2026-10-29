# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /app
# 先复制依赖文件，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 运行阶段
FROM alpine:3.19

# 安装 ca-certificates 确保能够正常调用外部 HTTPS API（OpenAI / 火山方舟 / Google 等）
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/main .
# 复制配置文件（兼容 config.yaml 或 config.example.yaml）
COPY manifest/config/config*.yaml manifest/config/
COPY workflows workflows/
RUN if [ ! -f manifest/config/config.yaml ]; then cp manifest/config/config.example.yaml manifest/config/config.yaml; fi

EXPOSE 8000
CMD ["./main"]
