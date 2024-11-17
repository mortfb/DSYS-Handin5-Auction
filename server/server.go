package main

import (
	proto "Auction/grpc"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type AuctionServer struct {
	proto.UnimplementedAuctionServer
	serverID           int
	serverPort         string
	highestBid         int
	highestBidder      string
	highestBidderID    int
	otherAuctionsConns []grpc.ClientConn
	otherAuctionPorts  []string
}

var finished bool = false

func main() {
	var input string
	log.Printf("Enter the server port (5050, 5051 or 5052)")
	fmt.Scan(&input)

	var auctionServer *AuctionServer
	if input == "5050" {
		auctionServer = &AuctionServer{
			serverID:          0,
			serverPort:        ":5050",
			highestBid:        0,
			highestBidder:     "",
			highestBidderID:   0,
			otherAuctionPorts: []string{":5051", ":5052"},
		}
	} else if input == "5051" {
		auctionServer = &AuctionServer{
			serverID:          1,
			serverPort:        ":5051",
			highestBid:        0,
			highestBidder:     "",
			highestBidderID:   0,
			otherAuctionPorts: []string{":5050", ":5052"},
		}
	} else if input == "5052" {
		auctionServer = &AuctionServer{
			serverID:          2,
			serverPort:        ":5052",
			highestBid:        0,
			highestBidder:     "",
			highestBidderID:   0,
			otherAuctionPorts: []string{":5050", ":5051"},
		}
	}

	go auctionServer.startServer()

	auctionServer.connectToOtherAuctions(auctionServer.otherAuctionPorts)

}

func (server *AuctionServer) startServer() {

	log.Printf("Server started")

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", server.serverPort)
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterAuctionServer(grpcServer, server)
	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}
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
	//Need a timer to check if the auction is finished

	if finished {
		log.Printf("Auction Result: %s gets the item for %d", Auction.highestBidder, Auction.highestBid)
		final_high_bid := Auction.highestBid
		final_winner := Auction.highestBidder
		//If we want to use this.
		final_winner_id := Auction.highestBidderID

		Auction.reset()
		return &proto.ResultResponse{Outcome: "Auction Result: " + string(final_winner_id) + " " + final_winner + " gets the item for " + string(final_high_bid), HighestBid: int32(final_high_bid), IsOver: true}, nil
	} else {
		return &proto.ResultResponse{Outcome: "Auction is still running, currently " + string(Auction.highestBidderID) + " " + Auction.highestBidder + " has the highest bid on " + string(Auction.highestBid), IsOver: false}, nil
	}
}

func (Auction *AuctionServer) reset() {
	Auction.highestBid = 0
	Auction.highestBidder = ""
	Auction.highestBidderID = 0
}

// updates the other acutions
func (Auction *AuctionServer) connectToOtherAuctions(otherAuctions []string) {
	for _, auction := range otherAuctions {
		if auction == Auction.serverPort {
			continue
		} else {
			//should store the connections in a list
			conn, err := grpc.Dial(auction, grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to auction %s: %v", auction, err)
				continue
			}
			Auction.otherAuctionsConns = append(Auction.otherAuctionsConns, *conn)
		}
	}
}
