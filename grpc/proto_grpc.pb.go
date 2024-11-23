// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package Auction

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AuctionClient is the client API for Auction service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuctionClient interface {
	Bid(ctx context.Context, in *BidRequest, opts ...grpc.CallOption) (*BidResponse, error)
	Result(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ResultResponse, error)
	SetID(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Client, error)
	SendElectionMessage(ctx context.Context, in *ElectionRequest, opts ...grpc.CallOption) (*ElectionResponse, error)
	SendUpdateBid(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*UpdateResponse, error)
	TestAlive(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	AuctionFinished(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type auctionClient struct {
	cc grpc.ClientConnInterface
}

func NewAuctionClient(cc grpc.ClientConnInterface) AuctionClient {
	return &auctionClient{cc}
}

func (c *auctionClient) Bid(ctx context.Context, in *BidRequest, opts ...grpc.CallOption) (*BidResponse, error) {
	out := new(BidResponse)
	err := c.cc.Invoke(ctx, "/Auction/bid", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionClient) Result(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ResultResponse, error) {
	out := new(ResultResponse)
	err := c.cc.Invoke(ctx, "/Auction/result", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionClient) SetID(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Client, error) {
	out := new(Client)
	err := c.cc.Invoke(ctx, "/Auction/setID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionClient) SendElectionMessage(ctx context.Context, in *ElectionRequest, opts ...grpc.CallOption) (*ElectionResponse, error) {
	out := new(ElectionResponse)
	err := c.cc.Invoke(ctx, "/Auction/sendElectionMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionClient) SendUpdateBid(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*UpdateResponse, error) {
	out := new(UpdateResponse)
	err := c.cc.Invoke(ctx, "/Auction/sendUpdateBid", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionClient) TestAlive(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/Auction/testAlive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionClient) AuctionFinished(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/Auction/auctionFinished", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuctionServer is the server API for Auction service.
// All implementations must embed UnimplementedAuctionServer
// for forward compatibility
type AuctionServer interface {
	Bid(context.Context, *BidRequest) (*BidResponse, error)
	Result(context.Context, *Empty) (*ResultResponse, error)
	SetID(context.Context, *Empty) (*Client, error)
	SendElectionMessage(context.Context, *ElectionRequest) (*ElectionResponse, error)
	SendUpdateBid(context.Context, *UpdateRequest) (*UpdateResponse, error)
	TestAlive(context.Context, *Empty) (*Empty, error)
	AuctionFinished(context.Context, *Empty) (*Empty, error)
	mustEmbedUnimplementedAuctionServer()
}

// UnimplementedAuctionServer must be embedded to have forward compatible implementations.
type UnimplementedAuctionServer struct {
}

func (UnimplementedAuctionServer) Bid(context.Context, *BidRequest) (*BidResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bid not implemented")
}
func (UnimplementedAuctionServer) Result(context.Context, *Empty) (*ResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Result not implemented")
}
func (UnimplementedAuctionServer) SetID(context.Context, *Empty) (*Client, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetID not implemented")
}
func (UnimplementedAuctionServer) SendElectionMessage(context.Context, *ElectionRequest) (*ElectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendElectionMessage not implemented")
}
func (UnimplementedAuctionServer) SendUpdateBid(context.Context, *UpdateRequest) (*UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendUpdateBid not implemented")
}
func (UnimplementedAuctionServer) TestAlive(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestAlive not implemented")
}
func (UnimplementedAuctionServer) AuctionFinished(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuctionFinished not implemented")
}
func (UnimplementedAuctionServer) mustEmbedUnimplementedAuctionServer() {}

// UnsafeAuctionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuctionServer will
// result in compilation errors.
type UnsafeAuctionServer interface {
	mustEmbedUnimplementedAuctionServer()
}

func RegisterAuctionServer(s grpc.ServiceRegistrar, srv AuctionServer) {
	s.RegisterService(&Auction_ServiceDesc, srv)
}

func _Auction_Bid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BidRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).Bid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/bid",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).Bid(ctx, req.(*BidRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auction_Result_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).Result(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/result",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).Result(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auction_SetID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).SetID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/setID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).SetID(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auction_SendElectionMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ElectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).SendElectionMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/sendElectionMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).SendElectionMessage(ctx, req.(*ElectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auction_SendUpdateBid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).SendUpdateBid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/sendUpdateBid",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).SendUpdateBid(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auction_TestAlive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).TestAlive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/testAlive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).TestAlive(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auction_AuctionFinished_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionServer).AuctionFinished(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction/auctionFinished",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionServer).AuctionFinished(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Auction_ServiceDesc is the grpc.ServiceDesc for Auction service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Auction_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Auction",
	HandlerType: (*AuctionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "bid",
			Handler:    _Auction_Bid_Handler,
		},
		{
			MethodName: "result",
			Handler:    _Auction_Result_Handler,
		},
		{
			MethodName: "setID",
			Handler:    _Auction_SetID_Handler,
		},
		{
			MethodName: "sendElectionMessage",
			Handler:    _Auction_SendElectionMessage_Handler,
		},
		{
			MethodName: "sendUpdateBid",
			Handler:    _Auction_SendUpdateBid_Handler,
		},
		{
			MethodName: "testAlive",
			Handler:    _Auction_TestAlive_Handler,
		},
		{
			MethodName: "auctionFinished",
			Handler:    _Auction_AuctionFinished_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/proto.proto",
}
