package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	pb "github.com/honhaoyuan0/FYP-Observability-on-SDN/Simplify_Demo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const name = "place_bid"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

func place_bid(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "place_bid")
	auction_round := 1 + rand.Intn(10)
	defer span.End()

	var msg string
	user := r.URL.Query().Get("user")
	bidStr := r.URL.Query().Get("bidValue")
	bandwidthStr := r.URL.Query().Get("bandwidthValue")

	if user == "" {
		user = "unknown_user"
	}

	var bidValue int
	if v, err := strconv.Atoi(bidStr); err != nil {
		logger.ErrorContext(ctx, "Invalid Bid Value",
			"error", err.Error(),
			"user", user,
		)
	} else {
		bidValue = v
	}

	var bandwidthValue int
	if v, err := strconv.Atoi(bandwidthStr); err != nil {
		logger.ErrorContext(ctx, "Invalid Bandwidth Value",
			"error", err.Error(),
			"user", user,
		)
	} else {
		bandwidthValue = v
	}

	logger.InfoContext(ctx,
		"User bids received successfully!",
		"auction_round", auction_round,
		"user", user,
		"bid_value", bidValue,
		"bandwidth_value", bandwidthValue,
	)

	msg = fmt.Sprintf("%s is placing a bid of %d for bandwidth of %d at auction round %d", user, bidValue, bandwidthValue, auction_round)
	logger.InfoContext(ctx, msg,
		"auction_round", auction_round,
		"user", user,
		"bid_value", bidValue,
		"bandwidth_value", bandwidthValue,
	)

	span.SetAttributes(
		attribute.Int("auction_round", auction_round),
		attribute.String("user", user),
		attribute.Int("bid_value", bidValue),
		attribute.Int("bandwidth_value", bandwidthValue),
	)

	// Resolve Auction Runner address from environment for Docker networking
	addr := os.Getenv("AUCTION_RUNNER_ADDR")
	if addr == "" {
		addr = "auction-runner:50051"
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// REPLACEMENT: Use WithStatsHandler instead of WithUnaryInterceptor
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	if err != nil {
		// log, set span error and return HTTP 500
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
		return
	}

	defer conn.Close()

	client := pb.NewAuctionClient(conn)
	grpcReq := &pb.BidRequest{
		User:           user,
		BidValue:       int32(bidValue),
		BandwidthValue: int32(bandwidthValue),
		AuctionRound:   int32(auction_round),
	}

	grpcCtx, cancel2 := context.WithTimeout(ctx, 3*time.Second)
	defer cancel2()

	grpcResp, err := client.PlaceBid(grpcCtx, grpcReq)
	if err != nil {
		// record/log error on span
		http.Error(w, "rpc error", http.StatusInternalServerError)
		return
	}

	// New child span for encoding & sending the HTTP response
	_, respSpan := tracer.Start(ctx, "encode-response")
	respSpan.SetAttributes(
		attribute.Bool("won", grpcResp.GetWon()),
		attribute.String("message", grpcResp.GetMessage()),
	)
	defer respSpan.End()

	// reply to original HTTP caller (JSON)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"won":     grpcResp.GetWon(),
		"message": grpcResp.GetMessage(),
	})
	logger.InfoContext(ctx,
		"Response received!",
		"auction_round", auction_round,
		"user", user,
		"bid_value", bidValue,
		"bandwidth_value", bandwidthValue,
	)
	respSpan.SetAttributes(
		attribute.Int("auction_round", auction_round),
		attribute.String("user", user),
		attribute.Int("bid_value", bidValue),
		attribute.Int("bandwidth_value", bandwidthValue),
	)

}
