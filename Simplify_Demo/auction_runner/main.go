package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"

	pb "github.com/honhaoyuan0/FYP-Observability-on-SDN/Simplify_Demo/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Setup OpenTelemetry
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return err
	}

	//Handle shutdown properly so nothing leaks
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	// Temporarily omit the OpenTelemetry gRPC interceptor to avoid
	// compatibility issues with the otelgrpc package/version.
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	pb.RegisterAuctionServer(grpcServer, NewAuctionServer())
	log.Println("Auction Runner gRPC listening :50051")

	// Serve in goroutine and stop gracefully on ctx cancellation.
	errCh := make(chan error, 1)
	go func() { errCh <- grpcServer.Serve(lis) }()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return nil
	case serveErr := <-errCh:
		return serveErr
	}
}
