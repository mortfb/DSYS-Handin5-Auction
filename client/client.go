package main

import (
	proto "Auction/grpc"
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
)

var thisClient *proto.Client

var currentBid int = 0

func main() {
	var name string
	var id int

	fmt.Println("Enter your name")
	fmt.Scan(&name)

	thisClient = &proto.Client{
		Name: name,
		ID:   int32(id),
	}

	// IMPORTANT: Connecting to the frontend, always on.
	conn, err := grpc.Dial(":5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	//connects the client to the frontendmanager
	manager := proto.NewAuctionClient(conn)

	//setting the client ID
	resID, _ := manager.SetID(context.Background(), &proto.Empty{})
	thisClient.ID = resID.ID

	var bid int

	for {
		fmt.Println("Enter the bid amount")
		fmt.Scan(&bid)

		if bid <= currentBid {
			fmt.Println("Bid must be higher than the current bid")
			continue
		} else {
			currentBid = bid
		}

		manager.PlaceBid(context.Background(), &proto.BidRequest{
			Amount: int32(bid),
			Client: thisClient,
		})

		res, err := manager.Result(context.Background(), &proto.Empty{})

		if err != nil {
			log.Fatalf("Failed to get result: %v", err)
		}

		if res.IsOver {
			//Do something when the auction is over
			fmt.Println(res.Outcome)
			break
		} else {
			fmt.Println(res.Outcome)
		}
	}

}
