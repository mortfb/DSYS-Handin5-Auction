package main

import (
	proto "Auction/grpc"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

// Acts as the middleman between the server and the clients.
// The manager gets a request from the client, which it then sends to the server.
type FrontEndManager struct {
	serverAdresses []string
	proto.UnimplementedAuctionServer
}

func main() {

	startManager()

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

}

// Fo when a client requests the result.
func (manager *FrontEndManager) result(ctx context.Context, req *proto.Empty) (*proto.ResultResponse, error) {

}
