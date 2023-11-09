// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.4
// source: proto/homework4proto.proto

package proto

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

// NodeServiceClient is the client API for NodeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeServiceClient interface {
	AnnounceConnection(ctx context.Context, in *ConnectionAnnouncement, opts ...grpc.CallOption) (*Confirmation, error)
	RequestAccess(ctx context.Context, in *AccessRequest, opts ...grpc.CallOption) (*AccessRequestResponse, error)
	AnnounceLeave(ctx context.Context, in *LeaveAnnouncement, opts ...grpc.CallOption) (*LeaveAnnouncementResponse, error)
	EnterCriticalSectionDirectly(ctx context.Context, in *AccessRequest, opts ...grpc.CallOption) (*Confirmation, error)
}

type nodeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeServiceClient(cc grpc.ClientConnInterface) NodeServiceClient {
	return &nodeServiceClient{cc}
}

func (c *nodeServiceClient) AnnounceConnection(ctx context.Context, in *ConnectionAnnouncement, opts ...grpc.CallOption) (*Confirmation, error) {
	out := new(Confirmation)
	err := c.cc.Invoke(ctx, "/homework4.NodeService/AnnounceConnection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) RequestAccess(ctx context.Context, in *AccessRequest, opts ...grpc.CallOption) (*AccessRequestResponse, error) {
	out := new(AccessRequestResponse)
	err := c.cc.Invoke(ctx, "/homework4.NodeService/RequestAccess", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) AnnounceLeave(ctx context.Context, in *LeaveAnnouncement, opts ...grpc.CallOption) (*LeaveAnnouncementResponse, error) {
	out := new(LeaveAnnouncementResponse)
	err := c.cc.Invoke(ctx, "/homework4.NodeService/AnnounceLeave", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) EnterCriticalSectionDirectly(ctx context.Context, in *AccessRequest, opts ...grpc.CallOption) (*Confirmation, error) {
	out := new(Confirmation)
	err := c.cc.Invoke(ctx, "/homework4.NodeService/EnterCriticalSectionDirectly", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeServiceServer is the server API for NodeService service.
// All implementations must embed UnimplementedNodeServiceServer
// for forward compatibility
type NodeServiceServer interface {
	AnnounceConnection(context.Context, *ConnectionAnnouncement) (*Confirmation, error)
	RequestAccess(context.Context, *AccessRequest) (*AccessRequestResponse, error)
	AnnounceLeave(context.Context, *LeaveAnnouncement) (*LeaveAnnouncementResponse, error)
	EnterCriticalSectionDirectly(context.Context, *AccessRequest) (*Confirmation, error)
	mustEmbedUnimplementedNodeServiceServer()
}

// UnimplementedNodeServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNodeServiceServer struct {
}

func (UnimplementedNodeServiceServer) AnnounceConnection(context.Context, *ConnectionAnnouncement) (*Confirmation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AnnounceConnection not implemented")
}
func (UnimplementedNodeServiceServer) RequestAccess(context.Context, *AccessRequest) (*AccessRequestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestAccess not implemented")
}
func (UnimplementedNodeServiceServer) AnnounceLeave(context.Context, *LeaveAnnouncement) (*LeaveAnnouncementResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AnnounceLeave not implemented")
}
func (UnimplementedNodeServiceServer) EnterCriticalSectionDirectly(context.Context, *AccessRequest) (*Confirmation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnterCriticalSectionDirectly not implemented")
}
func (UnimplementedNodeServiceServer) mustEmbedUnimplementedNodeServiceServer() {}

// UnsafeNodeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeServiceServer will
// result in compilation errors.
type UnsafeNodeServiceServer interface {
	mustEmbedUnimplementedNodeServiceServer()
}

func RegisterNodeServiceServer(s grpc.ServiceRegistrar, srv NodeServiceServer) {
	s.RegisterService(&NodeService_ServiceDesc, srv)
}

func _NodeService_AnnounceConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConnectionAnnouncement)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AnnounceConnection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/homework4.NodeService/AnnounceConnection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AnnounceConnection(ctx, req.(*ConnectionAnnouncement))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_RequestAccess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).RequestAccess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/homework4.NodeService/RequestAccess",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).RequestAccess(ctx, req.(*AccessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_AnnounceLeave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LeaveAnnouncement)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).AnnounceLeave(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/homework4.NodeService/AnnounceLeave",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).AnnounceLeave(ctx, req.(*LeaveAnnouncement))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeService_EnterCriticalSectionDirectly_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServiceServer).EnterCriticalSectionDirectly(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/homework4.NodeService/EnterCriticalSectionDirectly",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServiceServer).EnterCriticalSectionDirectly(ctx, req.(*AccessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeService_ServiceDesc is the grpc.ServiceDesc for NodeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "homework4.NodeService",
	HandlerType: (*NodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AnnounceConnection",
			Handler:    _NodeService_AnnounceConnection_Handler,
		},
		{
			MethodName: "RequestAccess",
			Handler:    _NodeService_RequestAccess_Handler,
		},
		{
			MethodName: "AnnounceLeave",
			Handler:    _NodeService_AnnounceLeave_Handler,
		},
		{
			MethodName: "EnterCriticalSectionDirectly",
			Handler:    _NodeService_EnterCriticalSectionDirectly_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/homework4proto.proto",
}
