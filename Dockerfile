# Description: Dockerfile for the API
FROM golang:1.22.5 AS base

# Arguments for the build
ARG TARGET_API
ARG API_PORT

# Environment variables
ENV TARGET_API=${TARGET_API}
ENV API_PORT=${API_PORT}

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Build the binaries

# Stage 1 - Build the binary
FROM base AS builder
WORKDIR /build
COPY . .
RUN go build ${TARGET_API}/cmd/app/main.go && \
    chmod +x main


# Stage 2: Compress the binary using UPX
FROM alpine AS upx
RUN apk add --no-cache upx
COPY --from=builder /build/main /upx/main
RUN upx --best --lzma /upx/main -o /upx/main_compressed

# MAIN
FROM scratch AS main
WORKDIR /app
COPY .env /app/.env
COPY --from=upx /upx/main_compressed /app/main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "./main" ]
EXPOSE ${API_PORT}
