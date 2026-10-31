FROM golang:1.26.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backend-service .

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/backend-service ./

EXPOSE 8989

CMD ["./backend-service"]