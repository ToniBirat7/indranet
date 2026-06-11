FROM golang:1.22-alpine AS builder
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /poc04 .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /poc04 /poc04
EXPOSE 8080
ENTRYPOINT ["/poc04"]
