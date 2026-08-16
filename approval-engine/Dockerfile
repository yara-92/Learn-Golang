# 多阶段构建：编译阶段用完整 Go 镜像，运行阶段用 scratch，最终镜像体积很小。
# 纯 Go SQLite 驱动（modernc.org/sqlite）不需要 CGO，天然适合这种极简镜像。

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM scratch
# schema.sql 已经通过 go:embed 编译进了二进制文件本身，运行阶段不需要单独拷贝。
COPY --from=builder /server /server
VOLUME ["/data"]
ENV APP_DB_PATH=/data/approval.db
ENV APP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/server"]
