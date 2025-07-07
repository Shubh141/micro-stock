FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o order .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/order .

EXPOSE 8081

CMD ["./order"]
