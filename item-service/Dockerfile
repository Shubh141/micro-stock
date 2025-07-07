FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o item .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/item .

EXPOSE 8080

CMD ["./item"]
