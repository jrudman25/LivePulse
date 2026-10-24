FROM golang:1.27.3-alpine3.24 AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/livepulse-server ./cmd/server

FROM alpine:3.24.3

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 livepulse \
    && adduser -S -D -H -u 10001 -G livepulse livepulse

WORKDIR /app
COPY --from=builder --chown=10001:10001 /out/livepulse-server /app/livepulse-server

ENV SERVER_PORT=8080
USER 10001:10001

EXPOSE 8080
ENTRYPOINT ["/app/livepulse-server"]
