# Stage 1: Build the Go backend
FROM golang:1.24.4-alpine AS backend-builder

WORKDIR /app

# 使用中国镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置 Go 代理为中国镜像
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG MODULE_NAME=github.com/xxcheng123/cloudpan189-share
ARG VAR_COMMIT
ARG VAR_BUILD_DATE
ARG VAR_GIT_SUMMARY
ARG VAR_GIT_BRANCH

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-X ${MODULE_NAME}/configs.Commit=${VAR_COMMIT} -X ${MODULE_NAME}/configs.BuildDate=${VAR_BUILD_DATE} -X ${MODULE_NAME}/configs.GitSummary=${VAR_GIT_SUMMARY} -X ${MODULE_NAME}/configs.GitBranch=${VAR_GIT_BRANCH}" -o /app/share ./cmd/main.go

# Stage 2: Build the Vue frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/fe

# 使用中国镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置 npm 为淘宝镜像
RUN npm config set registry https://registry.npmmirror.com

COPY fe/package.json fe/package-lock.json ./
RUN npm install

COPY fe/ ./

RUN npm run build

# Stage 3: Final image
FROM alpine:latest

WORKDIR /app

# 使用中国镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置时区
ENV TZ=Asia/Shanghai
ENV GIN_MODE=release
RUN apk add --no-cache ca-certificates tzdata

# Copy backend executable from backend-builder stage
COPY --from=backend-builder /app/share .

# Copy frontend static files from frontend-builder stage
COPY --from=frontend-builder /app/fe/dist ./fe/dist

# Copy configuration file
COPY etc/config.yaml ./etc/config.yaml

# Expose the port the application runs on (from config.yaml, default 12395)
EXPOSE 12395

# Command to run the application
CMD ["./share", "-config", "./etc/config.yaml"]