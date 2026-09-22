FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/backyard-proxy \
    .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/backyard-proxy /app/backyard-proxy

USER app

EXPOSE 8080

ENTRYPOINT ["/app/backyard-proxy"]