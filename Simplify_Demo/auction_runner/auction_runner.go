package main

import (
	"context"
	"fmt"

	pb "github.com/honhaoyuan0/FYP-Observability-on-SDN/Simplify_Demo/proto"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const name = "run-auction"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

type AuctionRunner struct {
	pb.UnimplementedAuctionServer
}

func NewAuctionServer() *AuctionRunner {
	return &AuctionRunner{}
}

func (s *AuctionRunner) PlaceBid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	ctx, span := tracer.Start(ctx, "run-auction")
	defer span.End()

	// Simple auction logic
	won := req.BidValue > 100
	msg := fmt.Sprintf("Bid Value insufficient you lost because your bid value %d is under 100", req.BidValue)
	if won {
		msg = fmt.Sprintf("%s bid value of %d accepted and you have won %d bandwidth for auction round %d with !", req.User, req.BidValue, req.BandwidthValue, req.AuctionRound)
	}

	logger.InfoContext(ctx, msg,
		"User", req.User,
		"Bid Value", int(req.BidValue),
		"Bid result", "WON",
		"Bandwidth Value", int(req.BandwidthValue),
		"Auction round", int(req.AuctionRound),
	)

	span.SetAttributes(
		attribute.String("User", req.User),
		attribute.Int("Bid Value", int(req.BidValue)),
		attribute.String("Bid result", "WON"),
		attribute.Int("Bandwidth Value", int(req.BandwidthValue)),
		attribute.Int("Auction round", int(req.AuctionRound)),
	)

	logger.InfoContext(ctx,
		"gRPC Responding",
		"User", req.User,
		"Bid Value", int(req.BidValue),
		"Bid result", "WON",
		"Bandwidth Value", int(req.BandwidthValue),
		"Auction round", int(req.AuctionRound),
	)
	return &pb.BidResponse{
		Won:     won,
		Message: msg,
	}, nil
}
