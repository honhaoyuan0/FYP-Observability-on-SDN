package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"

	pb "github.com/honhaoyuan0/FYP-Observability-on-SDN/Simplify_Demo/proto"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type server struct {
	pb.UnimplementedAuctionServer
}

func (s *server) PlaceBid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	won := (req.BidValue > 50)
	msg := "rejected"
	if won {
		msg = "accepted"
	}
	return &pb.BidResponse{Won: won, Message: msg}, nil
}

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
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	)

	pb.RegisterAuctionServer(grpcServer, &server{})
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
