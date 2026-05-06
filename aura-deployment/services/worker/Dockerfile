FROM golang:1.22-alpine AS builder
WORKDIR /build/aura-backend
COPY aura-backend/go.mod ./
COPY aura-backend/ ./
ARG CACHEBUST=
RUN printf '%s\n' "${CACHEBUST:-}" > /tmp/.aura-worker-cachebust \
  && CGO_ENABLED=0 go mod tidy \
  && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/aura ./cmd/aura-worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
USER nobody
COPY --from=builder /out/aura /usr/local/bin/aura
EXPOSE 8083
ENV HTTP_ADDR=:8083
ENTRYPOINT ["/usr/local/bin/aura"]
