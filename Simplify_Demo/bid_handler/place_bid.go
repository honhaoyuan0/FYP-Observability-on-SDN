package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	pb "github.com/honhaoyuan0/FYP-Observability-on-SDN/Simplify_Demo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
			attribute.String("error", err.Error()),
			attribute.String("user", user),
		)
	} else {
		bidValue = v
	}

	var bandwidthValue int
	if v, err := strconv.Atoi(bandwidthStr); err != nil {
		logger.ErrorContext(ctx, "Invalid Bandwidth Value",
			attribute.String("error", err.Error()),
			attribute.String("user", user),
		)
	} else {
		bandwidthValue = v
	}

	msg = fmt.Sprintf("%s is placing a bid of %d for bandwidth of %d at auction round %d", user, bidValue, bandwidthValue, auction_round)
	logger.InfoContext(ctx, msg,
		attribute.Int("auction_round", auction_round),
		attribute.String("user", user),
		attribute.Int("bid_value", bidValue),
		attribute.Int("bandwidth_value", bandwidthValue),
	)

	span.SetAttributes(
		attribute.Int("auction_round", auction_round),
		attribute.String("user", user),
		attribute.Int("bid_value", bidValue),
		attribute.Int("bandwidth_value", bandwidthValue),
	)

	ctxDial, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctxDial, "localhost:50051", 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	
	if err != nil {
		// log, set span error and return HTTP 500
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
		return
	}

	defer conn.Close()

	client := pb.NewAuctionClient(conn)
	grpcReq := &pb.BidRequest{
		User: user,
		BidValue: int32(bidValue),
		BandwidthValue: int32(bandwithValue)
		AuctionRound: int32(auction_round)
	}

	grpcCtx, cancel2 := context.WithTimeout(ctx, 3*time.Second)
	defer cancel2()

	grpcResp, err := client.PlaceBid(grpcCtx, grpcReq)
	if err != nil {
	// record/log error on span
	http.Error(w, "rpc error", http.StatusInternalServerError)
	return
	}

	// reply to original HTTP caller (JSON)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
	"won":    grpcResp.GetWon(),
	"message": grpcResp.GetMessage(),
	})

}
