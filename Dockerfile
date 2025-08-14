# 多阶段构建 Dockerfile
# 阶段1: 构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装前端依赖（包括开发依赖，因为构建需要）
RUN npm ci

# 复制前端源码
COPY frontend/ ./

# 构建前端（直接运行vite build，避免run-p依赖问题）
RUN npm run build-only

# 阶段2: 构建后端
FROM golang:1.23-alpine AS backend-builder

# 安装必要的工具
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY backend/ ./backend/

# 构建后端应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./backend/cmd/main.go

# 阶段3: 运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# 从构建阶段复制文件
COPY --from=backend-builder /app/main .
COPY --from=frontend-builder /app/frontend/dist ./backend/static/

# 复制配置文件
COPY .env.example .env

# 创建数据目录
RUN mkdir -p data logs

# 设置权限
RUN chmod +x main

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 启动应用
CMD ["./main"]
