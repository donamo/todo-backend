FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /todo-backend ./cmd/server

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /todo-backend /usr/local/bin/todo-backend
EXPOSE 3000
CMD ["todo-backend"]
