FROM golang:1.25-alpine AS builder

WORKDIR /app
# Copy proto module and service module
COPY proto/ ./proto/
COPY auction_runner/go.mod auction_runner/go.sum* ./auction_runner/

WORKDIR /app/auction_runner
RUN go mod download

# Copy auction_runner source code
COPY auction_runner/*.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o auction_runner .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auction_runner .

EXPOSE 50051

CMD ["./auction_runner"]
