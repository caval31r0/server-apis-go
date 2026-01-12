FROM golang:1.21-alpine AS builder

WORKDIR /app

# Dependências
COPY go.mod go.sum ./
RUN go mod download

# Código fonte
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/api/main.go

# Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /server .

EXPOSE 8080

CMD ["./server"]
