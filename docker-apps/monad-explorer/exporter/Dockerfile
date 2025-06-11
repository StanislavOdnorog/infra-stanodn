FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY exporter/main.go .
COPY exporter/go.mod .
COPY exporter/go.sum .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o monad-exporter main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/monad-exporter .

EXPOSE 8080
CMD ["./monad-exporter"] 