package main

import (
	proto "Auction/grpc"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

var nextServer proto.AuctionClient
var nextServerPort string

type AuctionServer struct {
	proto.UnimplementedAuctionServer
	serverID          int32
	serverPort        string
	highestBid        int
	highestBidder     string
	highestBidderID   int
	otherAuctionPorts []string
	updateCounter     int
	isLeader          bool
	leaderPort        string
	participate       bool
	numberClients     int
	leaderConn        *grpc.ClientConn
	leaderClient      proto.AuctionClient
}

var finished bool = false

func main() {
	var port string
	log.Printf("Enter the server port (5050, 5051 or 5052)")
	fmt.Scan(&port)

	var auctionServer *AuctionServer
	if port == "5050" {
		auctionServer = &AuctionServer{
			serverID:          0,
			serverPort:        ":5050",
			highestBid:        0,
			highestBidder:     "",
			highestBidderID:   0,
			otherAuctionPorts: []string{":5051", ":5052"},
			updateCounter:     0,
			isLeader:          false,
			participate:       false,
			numberClients:     0,
			leaderPort:        "",
		}
	} else if port == "5051" {
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
			numberClients:     0,
			leaderPort:        "",
		}
	} else if port == "5052" {
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
			numberClients:     0,
			leaderPort:        "",
		}
	}

	go auctionServer.startServer()

	auctionServer.setNextServer()

	if auctionServer.serverPort == ":5052" {
		auctionServer.startElection()
	}

	auctionServer.connectToLeader()

	waited := false
	var count int = 0

	for {
		if finished {
			count++
			if !auctionServer.isLeader && count == 1 {
				log.Printf("Auction Result: " + auctionServer.highestBidder + " gets the item for " + strconv.Itoa(auctionServer.highestBid))
			}
			if count == 1 {
				go func() {
					log.Printf("Auction is over")
					<-time.After(5 * time.Second)
					log.Printf("Shutting down")
					waited = true
				}()
			}

			if waited {
				break
			}
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

// updates a clients ID
func (Auction *AuctionServer) SetID(ctx context.Context, req *proto.Empty) (*proto.Client, error) {
	id := Auction.numberClients
	Auction.numberClients++
	return &proto.Client{ID: int32(id)}, nil
}

func (Auction *AuctionServer) Bid(ctx context.Context, req *proto.BidRequest) (*proto.BidResponse, error) {
	if Auction.isLeader {
		if !finished {
			Auction.updateCounter++
			log.Printf("Bid placed by %s for %d", req.Client.Name, req.Amount)
			if int(req.Amount) > Auction.highestBid {
				log.Printf("Highest bid is now %d by %s", req.Amount, req.Client.Name)
				Auction.highestBid = int(req.Amount)
				Auction.highestBidder = req.Client.Name
				Auction.highestBidderID = int(req.Client.ID)
			} else {
				return &proto.BidResponse{Message: "Bid Rejected"}, errors.New("bid must be higher than the current highest bid")
			}
			if Auction.updateCounter >= 100 {
				finished = true
			}

			Auction.sendUpdateToServers(ctx)
			return &proto.BidResponse{Message: "Bid Placed"}, nil

		} else {
			return &proto.BidResponse{Message: "The auction is over"}, errors.New("can't place bid when auction is over")
		}
	} else {
		if Auction.leaderConn == nil || Auction.leaderClient == nil {
			conn, err := grpc.Dial(Auction.leaderPort, grpc.WithInsecure())
			if err != nil {
				Auction.setNextServer()
				Auction.startElection()
				return &proto.BidResponse{Message: "Bid Rejected"}, err
			}
			Auction.leaderConn = conn
			Auction.leaderClient = proto.NewAuctionClient(conn)
			log.Printf("Connected to leader %s", Auction.leaderPort)
		}

		log.Printf("Bid placed by %s for %d, sending to leader", req.Client.Name, req.Amount)
		response, err := Auction.leaderClient.Bid(ctx, req)
		return response, err
	}
}

func (Auction *AuctionServer) Result(ctx context.Context, req *proto.Empty) (*proto.ResultResponse, error) {
	if Auction.isLeader {
		Auction.updateCounter++
		if Auction.updateCounter >= 100 {
			log.Printf("Auction Result: %s gets the item for %d", Auction.highestBidder, Auction.highestBid)
			time.Sleep(2 * time.Second)
			final_high_bid := Auction.highestBid
			final_winner := Auction.highestBidder
			finished = true

			Auction.NotifyAuctionFinished(context.Background())
			return &proto.ResultResponse{Outcome: "Auction Result: " + final_winner + " gets the item for " + strconv.Itoa(final_high_bid), HighestBid: int32(final_high_bid), IsOver: true}, nil
		} else {

			log.Println("updateCounter: " + strconv.Itoa(Auction.updateCounter))
			Auction.sendUpdateToServers(ctx)
			return &proto.ResultResponse{Outcome: "Auction is still running, currently " + Auction.highestBidder + " has the highest bid on " + strconv.Itoa(Auction.highestBid), IsOver: false}, nil
		}
	} else {
		if Auction.leaderConn == nil || Auction.leaderClient == nil {
			conn, err := grpc.Dial(Auction.leaderPort, grpc.WithInsecure())
			if err != nil {
				Auction.setNextServer()
				Auction.startElection()
				return nil, err
			}
			Auction.leaderConn = conn
			Auction.leaderClient = proto.NewAuctionClient(conn)
			log.Printf("Connected to leader %s", Auction.leaderPort)
		}

		return Auction.leaderClient.Result(ctx, req)
	}
}

// connects to the other auctionservers
func (Auction *AuctionServer) sendUpdateToServers(ctx context.Context) {
	for _, auction := range Auction.otherAuctionPorts {
		if auction == Auction.serverPort {
			continue
		} else {
			//maybe store the connections in a list
			conn, err := grpc.Dial(auction, grpc.WithTimeout(3*time.Second), grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Printf("Failed to connect to auction %s: %v", auction, err)
				continue
			}
			defer conn.Close()
			if conn != nil {
				node := proto.NewAuctionClient(conn)
				//Sends the updated bid to the other auctions
				node.SendUpdateBid(ctx, &proto.UpdateRequest{
					ServerID:      Auction.serverID,
					HighestBid:    int32(Auction.highestBid),
					HighestBidder: Auction.highestBidder,
					UpdateCounter: int32(Auction.updateCounter),
					NumberClients: int32(Auction.numberClients),
				})
			}
		}
	}
}

// this is only called by the leader
func (Auction *AuctionServer) SendUpdateBid(ctx context.Context, req *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	log.Printf("Update info received from leader")
	if int(req.UpdateCounter) > Auction.updateCounter {
		Auction.highestBid = int(req.HighestBid)
		Auction.highestBidder = req.HighestBidder
		Auction.updateCounter = int(req.UpdateCounter)
		Auction.numberClients = int(req.NumberClients)
		log.Printf("Highest bid is now %d by %s", Auction.highestBid, Auction.highestBidder)
		return &proto.UpdateResponse{
			Success: true,
		}, nil
	} else {
		return &proto.UpdateResponse{
			Success: false,
		}, nil
	}
}

func (Auction *AuctionServer) startElection() {
	log.Printf("Starting election")
	Auction.SendElectionMessage(context.Background(), &proto.ElectionRequest{ElectionPort: Auction.serverPort, LeaderID: -1, ServerID: Auction.serverID})
}

// Note: We need to figure out how we update the "next server" in the ring, in case of a server failure
func (Auction *AuctionServer) SendElectionMessage(ctx context.Context, req *proto.ElectionRequest) (*proto.ElectionResponse, error) {
	if nextServer == nil {
		Auction.setNextServer()
	} else if !Auction.checkLeader() {
		log.Printf("resetting next server")
		Auction.setNextServer()
	}
	log.Printf("hello?")
	if req.LeaderID == Auction.serverID {
		log.Printf("I won the Election %d", Auction.serverID)
		Auction.isLeader = true
		Auction.leaderPort = Auction.serverPort
		Auction.leaderConn = nil
		Auction.leaderClient = nil
	} else if req.LeaderID > Auction.serverID {
		log.Println("I am just sending this over to " + nextServerPort)
		Auction.leaderPort = req.ElectionPort
		_, err := nextServer.SendElectionMessage(ctx, &proto.ElectionRequest{
			ElectionPort: req.ElectionPort,
			ServerID:     req.ServerID,
			LeaderID:     req.LeaderID,
		})

		if err != nil {
			log.Fatalf("Failed to send token to next node: %v", err)
		}
		// Set the leader connection
		conn, err := grpc.Dial(req.ElectionPort, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to new leader: %v", err)
		}
		Auction.leaderConn = conn
		Auction.leaderClient = proto.NewAuctionClient(conn)
		log.Printf("Connected to new leader %s", req.ElectionPort)

	} else if !Auction.participate {
		log.Printf("I am the current winner")
		Auction.participate = true
		Auction.leaderPort = Auction.serverPort
		log.Printf("next server is: " + nextServerPort)
		_, err := nextServer.SendElectionMessage(ctx, &proto.ElectionRequest{
			ElectionPort: Auction.serverPort,
			ServerID:     Auction.serverID,
			LeaderID:     Auction.serverID,
		})
		if err != nil {
			log.Fatalf("Failed to send token to next node: %v", err)
		}
	}
	log.Printf("The leader is %s", Auction.leaderPort)
	return &proto.ElectionResponse{
		Success: true,
	}, nil
}

func (Auction *AuctionServer) setNextServer() {
	if Auction.serverPort == ":5050" {
		conn, err := grpc.Dial(":5051", grpc.WithTimeout(3*time.Second), grpc.WithInsecure(), grpc.WithBlock())
		nextServerPort = ":5051"
		if err != nil {
			log.Printf("Failed to connect to server: %v", err)
			conn, err = grpc.Dial(":5052", grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("Both other servers are down: %v", err)
			}
			nextServerPort = ":5052"
		}
		nextServer = proto.NewAuctionClient(conn)
	}

	if Auction.serverPort == ":5051" {
		log.Printf("please do something")
		conn, err := grpc.Dial(":5052", grpc.WithTimeout(3*time.Second), grpc.WithInsecure(), grpc.WithBlock())
		nextServerPort = ":5052"
		if err != nil {
			log.Printf("Failed to connect to server: %v", err)
			conn, err = grpc.Dial(":5050", grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("Both other servers are down: %v", err)
			}
			nextServerPort = ":5050"
		}
		nextServer = proto.NewAuctionClient(conn)
	}

	if Auction.serverPort == ":5052" {
		conn, err := grpc.Dial(":5050", grpc.WithTimeout(3*time.Second), grpc.WithInsecure(), grpc.WithBlock())
		nextServerPort = ":5050"
		if err != nil {
			log.Printf("Failed to connect to server: %v", err)
			conn, err = grpc.Dial(":5051", grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("Both other servers are down: %v", err)
			}
			nextServerPort = ":5051"
		}
		nextServer = proto.NewAuctionClient(conn)
	}
}

func (Auction *AuctionServer) checkLeader() bool {
	if Auction.isLeader {
		return true
	} else if Auction.leaderConn == nil {
		log.Printf("No leader connection")
		return false
	}

	log.Printf("Checking leader")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := Auction.leaderClient.TestAlive(ctx, &proto.Empty{})

	if err != nil {
		log.Printf("Leader is dead: %v", err)
		return false
	} else {
		log.Printf("Leader is alive")
		return true
	}
}

func (Auction *AuctionServer) connectToLeader() {
	checker := time.NewTicker(5 * time.Second)
	go func() {
		for range checker.C {
			if !Auction.isLeader && !finished {
				if Auction.leaderPort == "" {
					log.Printf("Leader port is not set")
					continue
				}
				if Auction.leaderConn == nil {
					log.Printf("Attempting to connect to leader at %s", Auction.leaderPort)
					conn, err := grpc.Dial(Auction.leaderPort, grpc.WithInsecure())
					if err != nil {
						log.Printf("Failed to connect to leader: %v", err)
						continue
					}
					Auction.leaderConn = conn
					Auction.leaderClient = proto.NewAuctionClient(conn)
					log.Printf("Connected to leader %s", Auction.leaderPort)

				} else if !Auction.checkLeader() {
					log.Printf("Leader is dead, starting election")
					if Auction.leaderConn != nil {
						Auction.leaderConn.Close()
						Auction.leaderConn = nil
						Auction.leaderClient = nil
					}
					Auction.setNextServer()
					Auction.startElection()
				}
			}
		}
	}()
}

func (Auction *AuctionServer) TestAlive(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, nil
}

func (Auction *AuctionServer) NotifyAuctionFinished(ctx context.Context) {
	for _, auction := range Auction.otherAuctionPorts {
		if auction == Auction.serverPort {
			continue
		} else {
			conn, err := grpc.Dial(auction, grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to auction %s: %v", auction, err)
				continue
			}
			defer conn.Close()

			node := proto.NewAuctionClient(conn)
			_, err = node.AuctionFinished(ctx, &proto.Empty{})
			if err != nil {
				log.Printf("Failed to notify auction %s: %v", auction, err)
			}
		}
	}
}

func (Auction *AuctionServer) AuctionFinished(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	finished = true
	return &proto.Empty{}, nil
}
