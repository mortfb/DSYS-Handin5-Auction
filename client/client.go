package main

import (
	proto "Auction/grpc"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"google.golang.org/grpc"
)

var thisClient *proto.Client

var currentBid int = 0

var serverPorts = [3]string{":5050", ":5051", ":5052"}

var auctionOver bool = false

func main() {
	var name string

	fmt.Println("Enter your name")
	fmt.Scan(&name)

	thisClient = &proto.Client{
		Name: name,
	}

	//Client connects to a random port, simulating front-end???????????.

	//adds the connection and the error early, so it can be changed later

	//connects the client to the frontendmanager
	node, err := connectToServer()
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	tmp, _ := node.SetID(context.Background(), &proto.Empty{})
	thisClient.ID = tmp.ID

	auctionOverchan := make(chan bool)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if node == nil {
				continue
			}
			res, err := node.Result(context.Background(), &proto.Empty{})
			if err != nil {
				log.Printf("Failed to get result: %v", err)
				node, _ = connectToServer()
				continue
			}

			if res.IsOver {
				log.Println(res.Outcome)
				auctionOverchan <- true
				break
			} else {
				log.Println(res.Outcome)
			}

		}
	}()

	var bid int

	go func() {
		for {
			if !auctionOver {
				fmt.Println("Enter the bid amount")
				fmt.Scan(&bid)

				if bid <= currentBid {
					fmt.Println("Bid must be higher than the current bid")
					continue
				} else {
					currentBid = bid
				}

				if node == nil {
					node, _ = connectToServer()
				}

				bidRes, erro := node.Bid(context.Background(), &proto.BidRequest{
					Amount: int32(currentBid),
					Client: thisClient,
				})

				if erro != nil {
					log.Printf("something went wrong with bidding %v", erro)
					log.Printf("Attempting to reconnect to server")
					node, err = connectToServer()
					if err != nil {
						log.Printf("Failed to connect to any server: %v", err)
					} else {
						bidRes, _ = node.Bid(context.Background(), &proto.BidRequest{
							Amount: int32(currentBid),
							Client: thisClient,
						})
					}
				}

				if bidRes != nil {
					log.Println(bidRes.Message)
				}

				if bid == -1 {
					res, err := node.Result(context.Background(), &proto.Empty{})

					if err != nil {
						log.Fatalf("Failed to get result: %v", err)
					}

					if res.IsOver {
						//Do something when the auction is over
						log.Println(res.Outcome)
						break
					} else {
						log.Println(res.Outcome)
					}
				} else if bid == -2 {
					break
				}
			}

		}

	}()

	<-auctionOverchan

	log.Printf("Auction is over")
}

func connectToServer() (proto.AuctionClient, error) {
	var randPort int = rand.IntN(len(serverPorts))
	//adds the connection and the error early, so it can be changed later
	conn, connErr := grpc.Dial(serverPorts[randPort], grpc.WithTimeout(3*time.Second), grpc.WithInsecure(), grpc.WithBlock())
	log.Printf("Attempting to connect to server on port: %s", serverPorts[randPort])
	if connErr == nil {
		log.Printf("Connected to  %s", serverPorts[randPort])
		client := proto.NewAuctionClient(conn)
		return client, nil
	}
	log.Printf("Failed to connect to server %s: %v", serverPorts[randPort], connErr)
	if connErr != nil {
		return connectToServer()
	}

	return nil, errors.New("failed to connect to any server")
}
