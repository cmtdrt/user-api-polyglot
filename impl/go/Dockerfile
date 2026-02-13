# Build
FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum* ./
COPY src ./src

RUN CGO_ENABLED=0 go build -o /user-api ./src

# Run
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /user-api /usr/local/bin/

EXPOSE 8080
CMD ["user-api"]
