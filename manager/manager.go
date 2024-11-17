package main

import (
	proto "Auction/grpc"
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// Acts as the middleman between the server and the clients.
// The manager gets a request from the client, which it then sends to the server.
type FrontEndManager struct {
	serverAdresses []string
	proto.UnimplementedAuctionServer
}

var numberOfClients int = 0
var mgLock sync.Mutex

func main() {
	startManager() //maybe this need to be a goroutine
}

//This function is called when the manager is started.

func startManager() {
	listen, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("Failed to listen on port 5000: %v", err)
	}

	frontEndManager := &FrontEndManager{
		serverAdresses: []string{":5050", ":5051", ":5052"},
	}

	grpcServer := grpc.NewServer()
	proto.RegisterAuctionServer(grpcServer, frontEndManager)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Called when client places a bid.
func (manager *FrontEndManager) placeBid(ctx context.Context, req *proto.BidRequest) (*proto.BidResponse, error) {
	//We might need to use locks
	//maybe make channels for the responses and errors
	mgLock.Lock()
	for _, address := range manager.serverAdresses {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to server %s: %v", address, err)
			continue
		}
		defer conn.Close()
		client := proto.NewAuctionClient(conn)

		response, err := client.PlaceBid(ctx, req)
		if err != nil {
			log.Printf("Failed to place bid on server %s: %v", address, err)
			continue
		}

		if response.Success {
			mgLock.Unlock()
			return response, nil
		}
	}
	mgLock.Unlock()

	// If no server accepted the bid, return a failure response
	return &proto.BidResponse{Message: "Bid Rejected by all servers", Success: false}, nil
}

// For when a client requests the result.
func (manager *FrontEndManager) result(ctx context.Context, req *proto.Empty) (*proto.ResultResponse, error) {
	//We might need to use locks
	//maybe make channels for the responses and errors
	mgLock.Lock()
	for _, address := range manager.serverAdresses {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to server %s: %v", address, err)
			continue
		}
		defer conn.Close()
		client := proto.NewAuctionClient(conn)

		response, err := client.Result(ctx, req)
		if err != nil {
			log.Printf("Failed to get result from server %s: %v", address, err)
			continue
		}
		mgLock.Unlock()
		return response, nil
	}
	mgLock.Unlock()
	// If no server accepted the bid, return a failure response
	return &proto.ResultResponse{Outcome: "No auction servers available"}, nil
}

func (manager *FrontEndManager) setID(ctx context.Context, req *proto.Empty) (*proto.Client, error) {
	mgLock.Lock()
	var id int = numberOfClients
	numberOfClients++
	mgLock.Unlock()
	return &proto.Client{ID: int32(id)}, nil
}
