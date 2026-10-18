FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/walletApp ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/walletApp .
EXPOSE 8080

CMD ["/app/walletApp"]