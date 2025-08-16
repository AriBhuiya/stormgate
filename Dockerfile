# ---------- build stage ----------
FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/stormgate ./cmd

# ---------- runtime stage ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
ENV CONFIG_PATH=/app/config.yaml
COPY --from=builder /out/stormgate /usr/local/bin/stormgate
EXPOSE 10000
USER app
ENTRYPOINT ["/usr/local/bin/stormgate"]