#Копирование и сборка
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

#Развёртывание и миграции
FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add tzdata
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./main"]
