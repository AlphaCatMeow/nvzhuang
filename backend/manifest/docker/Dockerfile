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

WORKDIR /app
COPY --from=builder /app/main .
# 复制配置（config.yaml 需在部署时通过环境变量注入 Key，或用 configmap 挂载）
COPY manifest/config/config.yaml manifest/config/config.yaml

EXPOSE 8000
CMD ["./main"]
