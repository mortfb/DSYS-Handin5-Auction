package main

import (
	proto "Auction/grpc"
	"log"
	"net"

	"google.golang.org/grpc"
)

type AuctionServer struct {
	proto.UnimplementedAuctionServer
}

func main() {
	startServer()
}

func startServer() {
	server := &AuctionServer{}

	log.Printf("Server started")

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterAuctionServer(grpcServer, server)
	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}
}
