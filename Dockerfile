FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o habit-tracker-bot .

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/habit-tracker-bot .

CMD ["./habit-tracker-bot"]