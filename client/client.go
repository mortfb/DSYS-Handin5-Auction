package main

import (
	proto "Auction/grpc"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var thisClient *proto.Client

var currentBid int = 0

var serverPorts = [3]string{":5050", ":5051", "5052"}

func main() {
	var name string
	var id int

	fmt.Println("Enter your name")
	fmt.Scan(&name)

	thisClient = &proto.Client{
		Name: name,
		ID:   int32(id),
	}

	//Client connects to a random port, simulating front-end???????????.
	var randPort int = rand.IntN(len(serverPorts))

	//adds the connection and the error early, so it can be changed later
	var conn *grpc.ClientConn
	var connErr error

	conn, connErr = grpc.Dial(serverPorts[randPort], grpc.WithInsecure())
	if connErr != nil {
		log.Fatalf("Failed to connect to server: %v", connErr)
	}

	//connects the client to the frontendmanager
	node := proto.NewAuctionClient(conn)
	/*
		//setting the client ID
		resID, _ := node.SetID(context.Background(), &proto.Empty{})
		thisClient.ID = resID.ID
	*/
	var bid int
	for {
		startTime := time.Now()

		state := conn.GetState()

		if state == connectivity.TransientFailure || state == connectivity.Shutdown && time.Since(startTime) >= 2*time.Second {
			log.Println("Connection lost!")

			randPort = rand.IntN(len(serverPorts))

			conn, connErr = grpc.Dial(serverPorts[randPort], grpc.WithInsecure())
			log.Printf("Connected to  %s", serverPorts[randPort])
			if connErr != nil {
				log.Fatalf("Failed to connect to server: %v", connErr)
			}

		}

		fmt.Println("Enter the bid amount")
		fmt.Scan(&bid)

		if bid <= currentBid {
			fmt.Println("Bid must be higher than the current bid")
			continue
		} else {
			currentBid = bid
		}

		bidRes, erro := node.PlaceBid(context.Background(), &proto.BidRequest{
			Amount: int32(bid),
			Client: thisClient,
		})

		if erro != nil {
			log.Printf("something went wrong with bidding %v", erro)
		}

		log.Println(bidRes.Message)

		res, err := node.Result(context.Background(), &proto.Empty{})

		if err != nil {
			log.Fatalf("Failed to get result: %v", err)
		}

		if res.IsOver {
			//Do something when the auction is over
			fmt.Println(res.Outcome)
		} else {
			fmt.Println(res.Outcome)
		}

		if bid == -1 {
			break
		}

	}

}
