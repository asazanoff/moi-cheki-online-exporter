# Стадия сборки
FROM golang:1.23.4-alpine3.21 AS builder

    WORKDIR /app

    COPY . .

    RUN go mod init qr-fns-exporter && go mod tidy

    RUN go build -o main .

# Стадия выполнения
FROM alpine:3.21.1

    WORKDIR /app

    COPY --from=builder /app/main .

    EXPOSE 8080

    CMD ["./main"]
