package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

const (
	KafkaTopic = "digest"
)

var (
	tracer = otel.Tracer("kafka-producer")
	logger = otelslog.NewLogger("kafka-producer")
)

func main() {
	ctx := context.Background()

	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		otelShutdown(ctx)
	}()

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_SERVER")})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	topic := KafkaTopic
	if err := produceFlows(ctx, p, topic, 10); err != nil {
		panic(err)
	}
}

func produceFlows(ctx context.Context, p *kafka.Producer, topic string, n int) error {
	ctx, span := tracer.Start(ctx, "produce-flows-batch")
	defer span.End()

	for i := 0; i < n; i++ {
		// Use parent context, don't create new Background()
		msgCtx, msgSpan := tracer.Start(ctx, "produce-flow-message")

		// Inject trace context into Kafka msg headers
		carrier := propagation.MapCarrier{}
		otel.GetTextMapPropagator().Inject(msgCtx, carrier)
		headers := make([]kafka.Header, 0, len(carrier))
		for k, v := range carrier {
			headers = append(headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}

		flow := Flow{
			IngressPort: int64(rand.Intn(4-1+1) + 1),
			EgressPort:  int64(rand.Intn(8-5+1) + 5),
			VlanID:      int64([]int{100, 200, 300}[rand.Intn(3)]),
			Bytes:       int64(rand.Intn(1000-500+1) + 500),
		}

		msgSpan.SetAttributes(
			attribute.Int("flow.id", i),
			attribute.Int64("flow.ingress_port", flow.IngressPort),
			attribute.Int64("flow.egress_port", flow.EgressPort),
			attribute.Int64("flow.vlan_id", flow.VlanID),
			attribute.Int64("flow.bytes", flow.Bytes),
		)

		payload, err := json.Marshal(flow)
		if err != nil {
			return err
		}

		key := uuid.NewString()
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(key),
			Value:          payload,
			Headers:        headers,
		}

		if err := p.Produce(msg, nil); err != nil {
			msgSpan.End()
			return err
		}
		logger.InfoContext(ctx, "Flow sent!",
			"flow.id", i,
			"flow.ingress_port", flow.IngressPort,
			"flow.egress_port", flow.EgressPort,
			"flow.vlan_id", flow.VlanID,
			"flow.bytes", flow.Bytes,
		)
		msgSpan.End()
	}

	// Wait for deliveries
	p.Flush(15000)
	return nil
}
