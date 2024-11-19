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

var serverPorts = [3]string{":5050", ":5051", ":5052"}

func main() {
	var name string

	fmt.Println("Enter your name")
	fmt.Scan(&name)

	thisClient = &proto.Client{
		Name: name,
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

	log.Printf("Connected to  %s", serverPorts[randPort])

	//connects the client to the frontendmanager
	node := proto.NewAuctionClient(conn)

	tmp, _ := node.SetID(context.Background(), &proto.Empty{})
	thisClient.ID = tmp.ID

	go func() {
		for {
			time.Sleep(5 * time.Second)
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
		}
	}()

	var bid int
	for {

		startTime := time.Now()

		//this should be done with a timer
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
			Amount: int32(currentBid),
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
			log.Println(res.Outcome)
		} else {
			log.Println(res.Outcome)
		}

		if bid == -1 {
			break
		}

	}

}
