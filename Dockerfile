FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/uno/*.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
ENTRYPOINT ["./server"]