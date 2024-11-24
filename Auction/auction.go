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
			highestBidder:     "NONE",
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
			highestBidder:     "NONE",
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
			highestBidder:     "NONE",
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
	var id int
	if Auction.isLeader {
		id = Auction.numberClients
		Auction.numberClients++
		return &proto.Client{ID: int32(id)}, nil
	} else {
		return Auction.leaderClient.SetID(ctx, req)
	}
}

func (Auction *AuctionServer) Bid(ctx context.Context, req *proto.BidRequest) (*proto.BidResponse, error) {
	if Auction.isLeader { //if node is the leader
		if !finished {
			Auction.updateCounter++
			log.Printf("Bid placed by %s for %d", req.Client.Name, req.Amount)
			if int(req.Amount) > Auction.highestBid {
				log.Printf("Highest bid is now %d by %s", req.Amount, req.Client.Name)
				Auction.highestBid = int(req.Amount)
				Auction.highestBidder = req.Client.Name
				Auction.highestBidderID = int(req.Client.ID)
			} else {
				Auction.sendUpdateToServers(ctx)
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
	} else { //if node is not the leader
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

			//Tekks the other servers that the auction is finished
			Auction.NotifyAuctionFinished(context.Background())
			return &proto.ResultResponse{Outcome: "Auction Result: " + final_winner + " gets the item for " + strconv.Itoa(final_high_bid), HighestBid: int32(final_high_bid), IsOver: true}, nil
		} else {

			if Auction.updateCounter%10 == 0 {
				log.Println("updateCounter: " + strconv.Itoa(Auction.updateCounter))
			}

			Auction.sendUpdateToServers(ctx)
			//returns the current highest bid
			return &proto.ResultResponse{Outcome: "Auction is still running, currently " + Auction.highestBidder + " has the highest bid on " + strconv.Itoa(Auction.highestBid), IsOver: false}, nil
		}
	} else { //if node is not the leader
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

		//passes the request to the leader, then returns the response to the client
		return Auction.leaderClient.Result(ctx, req)
	}
}

// connects to the other auctionservers
func (Auction *AuctionServer) sendUpdateToServers(ctx context.Context) {
	var tmpPorts []string

	for _, auction := range Auction.otherAuctionPorts {
		if auction == Auction.serverPort {
			continue
		} else {
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
					HighestBidID:  int32(Auction.highestBidderID),
				})
			}
			tmpPorts = append(tmpPorts, auction)
		}
	}
	Auction.otherAuctionPorts = tmpPorts
}

// sends the update bid to the other nodes
func (Auction *AuctionServer) SendUpdateBid(ctx context.Context, req *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	log.Printf("Update info received from leader")
	var prevBid int = Auction.highestBid

	if int(req.UpdateCounter) > Auction.updateCounter {
		Auction.highestBid = int(req.HighestBid)
		Auction.highestBidder = req.HighestBidder
		Auction.updateCounter = int(req.UpdateCounter)
		Auction.numberClients = int(req.NumberClients)
		Auction.highestBidderID = int(req.HighestBidID)
		if prevBid < Auction.highestBid {
			log.Printf("Highest bid is now %d by %s", Auction.highestBid, Auction.highestBidder)
		}
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
	Auction.SendElectionMessage(context.Background(), &proto.ElectionRequest{ElectionPort: Auction.serverPort, LeaderID: -1, ServerID: Auction.serverID}) //the server that calls the election is the first one to be "checked"
	//LeaderID is -1 to say that there is no "current leader" because no server has that ID
}

func (Auction *AuctionServer) SendElectionMessage(ctx context.Context, req *proto.ElectionRequest) (*proto.ElectionResponse, error) {
	if nextServer == nil { //this is to ensure that the connections between servers for the ring is established
		Auction.setNextServer()
	} else if !Auction.checkLeader() { //if the leader is dead, then the ring gets "reset" to exclude the leader
		Auction.setNextServer()
	}

	if req.LeaderID == Auction.serverID { //every server has been checked and its the leader's turn again for the second time. Nothing is larger than the leader
		log.Printf("I won the Election %d", Auction.serverID)
		Auction.isLeader = true
		Auction.leaderPort = Auction.serverPort
		//since this server is the leader, then it doesn't have to send anything to the leader. It also removes the connection from a previous (dead) leader
		Auction.leaderConn = nil
		Auction.leaderClient = nil
	} else if req.LeaderID > Auction.serverID { //the server is smaller than the leader so it passes the message it received to the next server
		log.Println("Sending election message to " + nextServerPort)
		Auction.leaderPort = req.ElectionPort
		_, err := nextServer.SendElectionMessage(ctx, &proto.ElectionRequest{ //message gets passed on here to the next server
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

	} else if !Auction.participate { //the server becomes the "current leading" server
		log.Printf("I am the current winner")
		Auction.participate = true
		Auction.leaderPort = Auction.serverPort
		log.Printf("next server is: " + nextServerPort)
		_, err := nextServer.SendElectionMessage(ctx, &proto.ElectionRequest{ //tells the next server that they are the current leader
			ElectionPort: Auction.serverPort,
			ServerID:     Auction.serverID,
			LeaderID:     Auction.serverID,
		})
		if err != nil {
			log.Fatalf("Failed to send token to next node: %v", err)
		}
	}
	log.Printf("The leader is %s", Auction.leaderPort) //this prints out when all recursive methods have been resolved, which happens when the leader is found
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

// Checks if the leader is still alive
func (Auction *AuctionServer) checkLeader() bool {
	if Auction.isLeader {
		return true
	} else if Auction.leaderConn == nil {
		log.Printf("No leader connection")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := Auction.leaderClient.TestAlive(ctx, &proto.Empty{})

	if err != nil {
		log.Printf("Leader is dead: %v", err)
		return false
	} else {
		return true
	}
}

// starts a ticker that checks if the leader is still alive on a set time interval, if not it starts an election
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
					//tries to connect to the leader, if no connection has been established
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
					//If leader is dead
					log.Printf("Leader is dead, starting election")
					if Auction.leaderConn != nil {
						//removes the connection to the now dead leader
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

// notify the other nodes that the auction is finished
func (Auction *AuctionServer) NotifyAuctionFinished(ctx context.Context) {
	var tmpPorts []string

	for _, auction := range Auction.otherAuctionPorts {
		if auction == Auction.serverPort {
			continue
		} else {
			conn, err := grpc.Dial(auction, grpc.WithTimeout(3*time.Second), grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to auction %s: %v", auction, err)
				//If we can't connect to the node, we assume its dead and remove it from the list
				continue
			}
			defer conn.Close()

			node := proto.NewAuctionClient(conn)
			_, err = node.AuctionFinished(ctx, &proto.Empty{})
			if err != nil {
				log.Printf("Failed to notify auction %s: %v", auction, err)
			}
		}

		//If we can connect to the node, we add it to the list
		tmpPorts = append(tmpPorts, auction)
	}
	//upades the list of other auction servers
	Auction.otherAuctionPorts = tmpPorts
}

func (Auction *AuctionServer) AuctionFinished(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	finished = true
	return &proto.Empty{}, nil
}
