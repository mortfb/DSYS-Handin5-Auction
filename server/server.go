package main

import (
	proto "Auction/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

var nextServer proto.AuctionClient

type AuctionServer struct {
	proto.UnimplementedAuctionServer
	serverID           int32
	serverPort         string
	highestBid         int
	highestBidder      string
	highestBidderID    int
	otherAuctionsConns []grpc.ClientConn
	otherAuctionPorts  []string
	updateCounter      int
	isLeader           bool
	leaderPort         string
	participate        bool
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
			updateCounter:     0,
			isLeader:          true,
			participate:       false,
		}
	} else if input == "5051" {
		auctionServer = &AuctionServer{
			serverID:          1,
			serverPort:        ":5051",
			highestBid:        0,
			highestBidder:     "",
			highestBidderID:   0,
			otherAuctionPorts: []string{":5050", ":5052"},
			updateCounter:     0,
			isLeader:          false,
			participate:       false,
		}
	} else if input == "5052" {
		auctionServer = &AuctionServer{
			serverID:          2,
			serverPort:        ":5052",
			highestBid:        0,
			highestBidder:     "",
			highestBidderID:   0,
			otherAuctionPorts: []string{":5050", ":5051"},
			updateCounter:     0,
			isLeader:          false,
			participate:       false,
		}

	}

	go auctionServer.startServer()

	auctionServer.connectToOtherAuctions(auctionServer.otherAuctionPorts)

	timer := time.NewTimer(10 * time.Second)

	for {
		select {
		case <-timer.C:
			finished = true
		default:

		}
	}

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
	if Auction.isLeader {
		if !finished {
			log.Printf("Bid placed by %s for %d", req.Client.Name, req.Amount)
			if int(req.Amount) > Auction.highestBid {
				Auction.updateCounter++
				Auction.highestBid = int(req.Amount)
				Auction.highestBidder = req.Client.Name
				Auction.highestBidderID = int(req.Client.ID)
				return &proto.BidResponse{Message: "Bid Placed"}, nil
			} else {
				return &proto.BidResponse{Message: "Bid Rejected"}, nil
			}
		} else {
			return &proto.BidResponse{Message: "The auction is over, bid rejected"}, nil
		}
	} else {
		//connects to the leader.
		conn, err := grpc.Dial(Auction.leaderPort, grpc.WithInsecure())
		defer conn.Close()

		//May need to have something here, that handles if the leader does not respond (ELECTION)
		if err != nil {
			Auction.startElection()
		}

		node := proto.NewAuctionClient(conn)
		response, err := node.PlaceBid(ctx, req)
		return response, err
	}

}

func (Auction *AuctionServer) result(ctx context.Context, req *proto.Empty) (*proto.ResultResponse, error) {
	//Need a timer to check if the auction is finished

	if finished {
		Auction.updateCounter++
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
	finished = false
}

// connects to the other auctionservers
func (Auction *AuctionServer) sendUpdatedBidToOtherAuctions() {
	for _, auction := range Auction.otherAuctionPorts {
		if auction == Auction.serverPort {
			continue
		} else {
			//should store the connections in a list
			conn, err := grpc.Dial(auction, grpc.WithInsecure())
			defer conn.Close()
			if err != nil {
				log.Printf("Failed to connect to auction %s: %v", auction, err)
				continue
			}
			node := proto.NewAuctionClient(conn)
			node.PlaceBid()
		}
	}
}

func (Auction *AuctionServer) startElection() {
	Auction.SendElectionMessage(context.Background(), &proto.ElectionRequest{
		ElectionMessage: Auction.serverID,
		ServerID:        Auction.serverID,
	})
}

func (Auction *AuctionServer) SendElectionMessage(ctx context.Context, req *proto.ElectionRequest) (*proto.ElectionResponse, error) {
	if req.ElectionMessage == Auction.serverID {
		Auction.isLeader = true
		nextServer.SendElectionMessage(ctx, &proto.ElectionRequest{
			ElectionMessage: Auction.serverID,
			ServerID:        Auction.serverID,
		})
	} else if req.ElectionMessage > Auction.serverID {
		nextServer.SendElectionMessage(ctx, req)
	} else if !Auction.participate {
		nextServer.SendElectionMessage(ctx, &proto.ElectionRequest{
			ElectionMessage: Auction.serverID,
			ServerID:        Auction.serverID,
		})
	}
	Auction.participate = true
	return &proto.ElectionResponse{
		Success: true,
	}, nil
}
