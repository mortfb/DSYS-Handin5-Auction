package main

import (
	proto "Auction/grpc"
)

func main() {
	var name string
	var id int

	client := &proto.Client{
		Name: name,
		ID:   int32(id),
	}

}
