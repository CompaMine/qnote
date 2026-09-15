# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/qnotes .

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates openssl \
  && adduser -D -u 1000 qnotes

WORKDIR /app
COPY --from=builder /out/qnotes /app/qnotes
COPY static /app/static
COPY entrypoint.sh /app/entrypoint.sh
RUN sed -i 's/\r$//' /app/entrypoint.sh \
  && chmod +x /app/entrypoint.sh \
  && mkdir -p /data \
  && chown -R qnotes:qnotes /app /data

ENV PORT=8443 \
    TLS_CERT=/data/cert.pem \
    TLS_KEY=/data/key.pem

VOLUME ["/data"]
EXPOSE 8443

# Start as root so volume perms / cert gen work, then drop to qnotes
USER root
ENTRYPOINT ["/app/entrypoint.sh"]
