FROM golang:1.25-alpine AS builder

WORKDIR /app
# Copy proto module and service module
COPY proto/ ./proto/
COPY bid_handler/go.mod bid_handler/go.sum* ./bid_handler/

WORKDIR /app/bid_handler
RUN go mod download

# Copy bid_handler source code
COPY bid_handler/*.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o bid_handler .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bid_handler .

EXPOSE 8080

CMD ["./bid_handler"]
