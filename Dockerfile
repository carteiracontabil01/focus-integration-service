# ---------- build ----------
FROM golang:1.24-alpine AS builder
WORKDIR /app

# evita problemas com toolchain em ambientes com GOTOOLCHAIN=local
ENV GOTOOLCHAIN=auto

# dependências primeiro p/ cache
COPY go.mod go.sum ./
RUN go mod download

# código
COPY . .

# build do entrypoint
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/api

# ---------- run ----------
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 8082
ENTRYPOINT ["/app/server"]


