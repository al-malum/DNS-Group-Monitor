FROM golang:1.21.3 AS builder
WORKDIR /build
COPY .  .
RUN go build -o dns-group-monitor \
    -ldflags "-X main.desiredPathPid=/run/dns-exporter.pid" \
    cmd/pdns/main.go
RUN ls -l /build

FROM golang:alpine AS runner
WORKDIR /app
RUN apk add gcompat
COPY --from=builder /build/dns-group-monitor /app/
COPY --from=builder /build/config.json /app/
CMD ["/app/dns-group-monitor", "-c", "/app/config.json"]