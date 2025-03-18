FROM golang:1.24.1 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o gateway ./main.go

# Финальный образ
FROM debian:bookworm


WORKDIR /app

COPY --from=builder /app/gateway /app/gateway

CMD ["/app/gateway"]
