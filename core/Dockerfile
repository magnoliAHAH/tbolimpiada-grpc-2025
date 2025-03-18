FROM golang:1.24.1 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o core ./main.go

# Финальный образ
FROM debian:bookworm

WORKDIR /app

COPY --from=builder /app/core /app/core

CMD ["/app/core"]
