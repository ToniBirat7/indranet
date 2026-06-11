FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /poc02 .

FROM scratch
COPY --from=builder /poc02 /poc02
EXPOSE 8765
ENTRYPOINT ["/poc02"]
