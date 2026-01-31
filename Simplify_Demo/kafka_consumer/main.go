package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

var (
	tracer = otel.Tracer("kafka-consumer")
	logger = otelslog.NewLogger("kafka-consumer")
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry SDK
	shutdown, err := setupOTelSDK(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down OTel SDK: %v", err)
		}
	}()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_SERVER"),
		"group.id":          "telemetry-service",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.SubscribeTopics([]string{"digest"}, nil)
	if err != nil {
		panic(err)
	}

	run := true
	for run {
		msg, err := c.ReadMessage(time.Second)

		if err != nil {
			if !err.(kafka.Error).IsTimeout() {
				log.Printf("Consumer error: %v", err)
			}
			continue
		}

		// Extract trace context from message headers
		carrier := propagation.MapCarrier{}
		for _, h := range msg.Headers {
			carrier[h.Key] = string(h.Value)
		}

		ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
		ctx, span := tracer.Start(ctx, "consume-flow")

		// Unmarshal the Flow data
		var flow Flow
		if err := json.Unmarshal(msg.Value, &flow); err != nil {
			logger.InfoContext(ctx, "Failed to unmarshal flow",
				"error", err,
			)
			span.End()
			continue
		}

		// Add flow data as span attributes
		span.SetAttributes(
			attribute.Int64("flow.ingress_port", flow.IngressPort),
			attribute.Int64("flow.egress_port", flow.EgressPort),
			attribute.Int64("flow.vlan_id", flow.VlanID),
			attribute.Int64("flow.bytes", flow.Bytes),
		)

		logger.InfoContext(ctx, "Flow consumed and processed",
			"flow.ingress_port", flow.IngressPort,
			"flow.egress_port", flow.EgressPort,
			"flow.vlan_id", flow.VlanID,
			"flow.bytes", flow.Bytes,
		)

		span.End()
	}
}
