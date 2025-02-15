
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/app

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080

ENV DB_DSN="postgres://postgres:password@db:5432/avito?sslmode=disable"

CMD ["./server"]