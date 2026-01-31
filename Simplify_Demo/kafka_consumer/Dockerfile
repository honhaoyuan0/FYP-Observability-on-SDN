FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY kafka_consumer/go.mod kafka_consumer/go.sum* ./kafka_consumer/

WORKDIR /app/kafka_consumer
RUN go mod download

COPY kafka_consumer/*.go ./

# Resolve dependencies
RUN go mod tidy

# Build with CGO for librdkafka
RUN go build -o consumer .

# Final stage
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    librdkafka1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/kafka_consumer/consumer .

ENV KAFKA_SERVER=kafka:9092
CMD ["./consumer"]
