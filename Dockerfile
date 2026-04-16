# ============================================
# K8s-Manager All-in-One Dockerfile
# Frontend (Nginx) + Backend (Go) in one image
# ============================================

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY version.sh ./
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY frontend/ ./
RUN sh -c '. ./version.sh && npm pkg set version="$VERSION"'
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /backend
ENV GOPROXY=https://goproxy.cn,direct
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o k8s-manager ./cmd/server

# Stage 3: Runtime
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata nginx \
    curl busybox-extras bind-tools bash iproute2 && \
    mkdir -p /run/nginx

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /backend/k8s-manager .

# Copy frontend static files
COPY --from=frontend-builder /frontend/dist /usr/share/nginx/html

# Copy nginx config
COPY frontend/nginx.conf /etc/nginx/http.d/default.conf

# Copy default backend config
COPY backend/configs/config.yaml /app/configs/config.yaml

# Create startup script
RUN printf '#!/bin/sh\nnginx\nexec /app/k8s-manager --config=${CONFIG_PATH:-/app/configs/config.yaml}\n' > /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 8080

ENTRYPOINT ["/app/start.sh"]
