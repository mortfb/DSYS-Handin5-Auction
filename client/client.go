package main

import (
	proto "Auction/grpc"
	"log"

	"google.golang.org/grpc"
)

var thisClient *proto.Client

func main() {
	var name string
	var id int

	thisClient = &proto.Client{
		Name: name,
		ID:   int32(id),
	}

	// IMPORTANT: Connecting to the frontend.
	conn, err := grpc.Dial(":5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	client := proto.NewAuctionClient(conn)

	for {
		//maybe placing bids should be done as manual input in the terminal

	}

}
