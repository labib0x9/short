# ---- build ----
FROM golang:1.26-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/short ./cmd/short
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/bootstrap ./cmd/bootstrap

# ---- runtime ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/short ./short
COPY --from=builder /out/bootstrap ./bootstrap
COPY migrations ./migrations
COPY static ./static

EXPOSE 3000

ENTRYPOINT ["/app/short"]