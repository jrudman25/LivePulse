# ==========================================
# 1. BUILD STAGE
# ==========================================
FROM golang:alpine AS builder

# Set the working directory directly inside the container
WORKDIR /build

# Pull in certificates (required for secure HTTPS fetch calls to Ticketmaster)
RUN apk update && apk add --no-cache ca-certificates tzdata

# Copy the Go Modules manifests (this enables Docker caching so builds are blazing fast)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entirety of your backend source code
COPY . .

# Strip out unnecessary debug symbols and statically compile the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /livepulse-server ./cmd/server

# ==========================================
# 2. DEPLOYMENT STAGE
# ==========================================
FROM alpine:latest

WORKDIR /app

# Copy the generated TLS certs from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the raw compiled Go executable from the builder down to the final container
COPY --from=builder /livepulse-server /app/livepulse-server

# Tell Northflank exactly which port our server listens to natively!
EXPOSE 8080

# Fire up the engine automatically!
CMD ["/app/livepulse-server"]
