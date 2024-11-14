package main

import (
	proto "Auction/grpc"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type AuctionServer struct {
	proto.UnimplementedAuctionServer
	serverID        int
	serverPort      string
	highestBid      int
	highestBidder   string
	highestBidderID int
}

func main() {
	startServer()
}

func (Auction *AuctionServer) placeBid(ctx context.Context, req *proto.BidRequest) (*proto.BidResponse, error) {
	log.Printf("Bid placed by %s for %d", req.Client.Name, req.Amount)
	if int(req.Amount) > Auction.highestBid {
		Auction.highestBid = int(req.Amount)
		Auction.highestBidder = req.Client.Name
		Auction.highestBidderID = int(req.Client.ID)
		return &proto.BidResponse{Message: "Bid Placed"}, nil
	} else {
		return &proto.BidResponse{Message: "Bid Rejected"}, nil
	}
}

func (Auction *AuctionServer) result(ctx context.Context, req *proto.Empty) (*proto.ResultResponse, error) {
	log.Printf("Auction Result: %s gets the item for %d", Auction.highestBidder, Auction.highestBid)

	return &proto.ResultResponse{Outcome: "Auction Result: " + Auction.highestBidder + " gets the item for " + string(Auction.highestBid), HighestBid: int32(Auction.highestBid)}, nil
}

func startServer() {
	server := &AuctionServer{}

	log.Printf("Server started")

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterAuctionServer(grpcServer, server)
	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}
}
