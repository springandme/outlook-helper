# 多阶段构建 Dockerfile
# 阶段1: 构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装前端依赖（包括开发依赖，因为构建需要）
RUN npm ci --no-audit --no-fund

# 显示npm和node版本信息
RUN node --version && npm --version

# 复制前端源码
COPY frontend/ ./

# 接受构建参数
ARG VITE_API_BASE_URL=/api
ARG VITE_APP_TITLE=Outlook取件助手
ARG VITE_APP_VERSION=1.0.0

# 设置前端构建环境变量
ENV NODE_ENV=production
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_APP_TITLE=$VITE_APP_TITLE
ENV VITE_APP_VERSION=$VITE_APP_VERSION

# 显示环境变量（用于调试）
RUN echo "🔧 构建环境变量:" && \
    echo "NODE_ENV=$NODE_ENV" && \
    echo "VITE_API_BASE_URL=$VITE_API_BASE_URL" && \
    echo "VITE_APP_TITLE=$VITE_APP_TITLE"

# 清理可能的缓存并构建前端
RUN echo "🧹 清理缓存..." && \
    npm cache clean --force && \
    rm -rf node_modules/.cache dist && \
    echo "🔨 开始构建前端..." && \
    npm run build-only

# 内嵌验证构建结果
RUN echo "🔍 验证构建结果..." && \
    # 检查dist目录是否存在 \
    if [ ! -d "dist" ]; then echo "❌ 错误：dist目录不存在"; exit 1; fi && \
    # 检查index.html是否存在 \
    if [ ! -f "dist/index.html" ]; then echo "❌ 错误：index.html文件不存在"; exit 1; fi && \
    # 检查是否有JS文件 \
    if [ $(find dist -name "*.js" | wc -l) -eq 0 ]; then echo "❌ 错误：没有找到JS文件"; exit 1; fi && \
    # 检查JS文件中是否包含硬编码的localhost:8080 \
    if grep -r "localhost:8080" dist/ 2>/dev/null; then echo "❌ 错误：发现硬编码的localhost:8080"; exit 1; fi && \
    # 显示构建产物信息 \
    echo "✅ 构建验证通过！" && \
    echo "📦 构建产物大小: $(du -sh dist | cut -f1)" && \
    echo "📁 主要文件:" && \
    ls -la dist/

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
