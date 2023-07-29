// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: prom/v1/node.proto

package v1

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

// NodeClient is the client API for Node service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeClient interface {
	CreateNode(ctx context.Context, in *CreateNodeRequest, opts ...grpc.CallOption) (*CreateNodeReply, error)
	UpdateNode(ctx context.Context, in *UpdateNodeRequest, opts ...grpc.CallOption) (*UpdateNodeReply, error)
	DeleteNode(ctx context.Context, in *DeleteNodeRequest, opts ...grpc.CallOption) (*DeleteNodeReply, error)
	GetNode(ctx context.Context, in *GetNodeRequest, opts ...grpc.CallOption) (*GetNodeReply, error)
	ListNode(ctx context.Context, in *ListNodeRequest, opts ...grpc.CallOption) (*ListNodeReply, error)
}

type nodeClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeClient(cc grpc.ClientConnInterface) NodeClient {
	return &nodeClient{cc}
}

func (c *nodeClient) CreateNode(ctx context.Context, in *CreateNodeRequest, opts ...grpc.CallOption) (*CreateNodeReply, error) {
	out := new(CreateNodeReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.Node/CreateNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) UpdateNode(ctx context.Context, in *UpdateNodeRequest, opts ...grpc.CallOption) (*UpdateNodeReply, error) {
	out := new(UpdateNodeReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.Node/UpdateNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) DeleteNode(ctx context.Context, in *DeleteNodeRequest, opts ...grpc.CallOption) (*DeleteNodeReply, error) {
	out := new(DeleteNodeReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.Node/DeleteNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) GetNode(ctx context.Context, in *GetNodeRequest, opts ...grpc.CallOption) (*GetNodeReply, error) {
	out := new(GetNodeReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.Node/GetNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeClient) ListNode(ctx context.Context, in *ListNodeRequest, opts ...grpc.CallOption) (*ListNodeReply, error) {
	out := new(ListNodeReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.Node/ListNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeServer is the server API for Node service.
// All implementations must embed UnimplementedNodeServer
// for forward compatibility
type NodeServer interface {
	CreateNode(context.Context, *CreateNodeRequest) (*CreateNodeReply, error)
	UpdateNode(context.Context, *UpdateNodeRequest) (*UpdateNodeReply, error)
	DeleteNode(context.Context, *DeleteNodeRequest) (*DeleteNodeReply, error)
	GetNode(context.Context, *GetNodeRequest) (*GetNodeReply, error)
	ListNode(context.Context, *ListNodeRequest) (*ListNodeReply, error)
	mustEmbedUnimplementedNodeServer()
}

// UnimplementedNodeServer must be embedded to have forward compatible implementations.
type UnimplementedNodeServer struct {
}

func (UnimplementedNodeServer) CreateNode(context.Context, *CreateNodeRequest) (*CreateNodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateNode not implemented")
}
func (UnimplementedNodeServer) UpdateNode(context.Context, *UpdateNodeRequest) (*UpdateNodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNode not implemented")
}
func (UnimplementedNodeServer) DeleteNode(context.Context, *DeleteNodeRequest) (*DeleteNodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNode not implemented")
}
func (UnimplementedNodeServer) GetNode(context.Context, *GetNodeRequest) (*GetNodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNode not implemented")
}
func (UnimplementedNodeServer) ListNode(context.Context, *ListNodeRequest) (*ListNodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListNode not implemented")
}
func (UnimplementedNodeServer) mustEmbedUnimplementedNodeServer() {}

// UnsafeNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeServer will
// result in compilation errors.
type UnsafeNodeServer interface {
	mustEmbedUnimplementedNodeServer()
}

func RegisterNodeServer(s grpc.ServiceRegistrar, srv NodeServer) {
	s.RegisterService(&Node_ServiceDesc, srv)
}

func _Node_CreateNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).CreateNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.Node/CreateNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).CreateNode(ctx, req.(*CreateNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_UpdateNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).UpdateNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.Node/UpdateNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).UpdateNode(ctx, req.(*UpdateNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_DeleteNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).DeleteNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.Node/DeleteNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).DeleteNode(ctx, req.(*DeleteNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_GetNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).GetNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.Node/GetNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).GetNode(ctx, req.(*GetNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Node_ListNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).ListNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.Node/ListNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).ListNode(ctx, req.(*ListNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Node_ServiceDesc is the grpc.ServiceDesc for Node service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Node_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.prom.v1.Node",
	HandlerType: (*NodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateNode",
			Handler:    _Node_CreateNode_Handler,
		},
		{
			MethodName: "UpdateNode",
			Handler:    _Node_UpdateNode_Handler,
		},
		{
			MethodName: "DeleteNode",
			Handler:    _Node_DeleteNode_Handler,
		},
		{
			MethodName: "GetNode",
			Handler:    _Node_GetNode_Handler,
		},
		{
			MethodName: "ListNode",
			Handler:    _Node_ListNode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "prom/v1/node.proto",
}