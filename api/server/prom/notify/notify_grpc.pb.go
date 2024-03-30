// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.4
// source: server/prom/notify/notify.proto

package notify

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

const (
	Notify_CreateNotify_FullMethodName = "/api.server.prom.notify.Notify/CreateNotify"
	Notify_UpdateNotify_FullMethodName = "/api.server.prom.notify.Notify/UpdateNotify"
	Notify_DeleteNotify_FullMethodName = "/api.server.prom.notify.Notify/DeleteNotify"
	Notify_GetNotify_FullMethodName    = "/api.server.prom.notify.Notify/GetNotify"
	Notify_ListNotify_FullMethodName   = "/api.server.prom.notify.Notify/ListNotify"
	Notify_SelectNotify_FullMethodName = "/api.server.prom.notify.Notify/SelectNotify"
)

// NotifyClient is the client API for Notify service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NotifyClient interface {
	// 创建通知对象
	CreateNotify(ctx context.Context, in *CreateNotifyRequest, opts ...grpc.CallOption) (*CreateNotifyReply, error)
	// 更新通知对象
	UpdateNotify(ctx context.Context, in *UpdateNotifyRequest, opts ...grpc.CallOption) (*UpdateNotifyReply, error)
	// 删除通知对象
	DeleteNotify(ctx context.Context, in *DeleteNotifyRequest, opts ...grpc.CallOption) (*DeleteNotifyReply, error)
	// 获取通知对象详情
	GetNotify(ctx context.Context, in *GetNotifyRequest, opts ...grpc.CallOption) (*GetNotifyReply, error)
	// 获取通知对象列表
	ListNotify(ctx context.Context, in *ListNotifyRequest, opts ...grpc.CallOption) (*ListNotifyReply, error)
	// 获取通知对象列表(用于下拉选择)
	SelectNotify(ctx context.Context, in *SelectNotifyRequest, opts ...grpc.CallOption) (*SelectNotifyReply, error)
}

type notifyClient struct {
	cc grpc.ClientConnInterface
}

func NewNotifyClient(cc grpc.ClientConnInterface) NotifyClient {
	return &notifyClient{cc}
}

func (c *notifyClient) CreateNotify(ctx context.Context, in *CreateNotifyRequest, opts ...grpc.CallOption) (*CreateNotifyReply, error) {
	out := new(CreateNotifyReply)
	err := c.cc.Invoke(ctx, Notify_CreateNotify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyClient) UpdateNotify(ctx context.Context, in *UpdateNotifyRequest, opts ...grpc.CallOption) (*UpdateNotifyReply, error) {
	out := new(UpdateNotifyReply)
	err := c.cc.Invoke(ctx, Notify_UpdateNotify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyClient) DeleteNotify(ctx context.Context, in *DeleteNotifyRequest, opts ...grpc.CallOption) (*DeleteNotifyReply, error) {
	out := new(DeleteNotifyReply)
	err := c.cc.Invoke(ctx, Notify_DeleteNotify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyClient) GetNotify(ctx context.Context, in *GetNotifyRequest, opts ...grpc.CallOption) (*GetNotifyReply, error) {
	out := new(GetNotifyReply)
	err := c.cc.Invoke(ctx, Notify_GetNotify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyClient) ListNotify(ctx context.Context, in *ListNotifyRequest, opts ...grpc.CallOption) (*ListNotifyReply, error) {
	out := new(ListNotifyReply)
	err := c.cc.Invoke(ctx, Notify_ListNotify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notifyClient) SelectNotify(ctx context.Context, in *SelectNotifyRequest, opts ...grpc.CallOption) (*SelectNotifyReply, error) {
	out := new(SelectNotifyReply)
	err := c.cc.Invoke(ctx, Notify_SelectNotify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotifyServer is the server API for Notify service.
// All implementations must embed UnimplementedNotifyServer
// for forward compatibility
type NotifyServer interface {
	// 创建通知对象
	CreateNotify(context.Context, *CreateNotifyRequest) (*CreateNotifyReply, error)
	// 更新通知对象
	UpdateNotify(context.Context, *UpdateNotifyRequest) (*UpdateNotifyReply, error)
	// 删除通知对象
	DeleteNotify(context.Context, *DeleteNotifyRequest) (*DeleteNotifyReply, error)
	// 获取通知对象详情
	GetNotify(context.Context, *GetNotifyRequest) (*GetNotifyReply, error)
	// 获取通知对象列表
	ListNotify(context.Context, *ListNotifyRequest) (*ListNotifyReply, error)
	// 获取通知对象列表(用于下拉选择)
	SelectNotify(context.Context, *SelectNotifyRequest) (*SelectNotifyReply, error)
	mustEmbedUnimplementedNotifyServer()
}

// UnimplementedNotifyServer must be embedded to have forward compatible implementations.
type UnimplementedNotifyServer struct {
}

func (UnimplementedNotifyServer) CreateNotify(context.Context, *CreateNotifyRequest) (*CreateNotifyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateNotify not implemented")
}
func (UnimplementedNotifyServer) UpdateNotify(context.Context, *UpdateNotifyRequest) (*UpdateNotifyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNotify not implemented")
}
func (UnimplementedNotifyServer) DeleteNotify(context.Context, *DeleteNotifyRequest) (*DeleteNotifyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNotify not implemented")
}
func (UnimplementedNotifyServer) GetNotify(context.Context, *GetNotifyRequest) (*GetNotifyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNotify not implemented")
}
func (UnimplementedNotifyServer) ListNotify(context.Context, *ListNotifyRequest) (*ListNotifyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListNotify not implemented")
}
func (UnimplementedNotifyServer) SelectNotify(context.Context, *SelectNotifyRequest) (*SelectNotifyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SelectNotify not implemented")
}
func (UnimplementedNotifyServer) mustEmbedUnimplementedNotifyServer() {}

// UnsafeNotifyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NotifyServer will
// result in compilation errors.
type UnsafeNotifyServer interface {
	mustEmbedUnimplementedNotifyServer()
}

func RegisterNotifyServer(s grpc.ServiceRegistrar, srv NotifyServer) {
	s.RegisterService(&Notify_ServiceDesc, srv)
}

func _Notify_CreateNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServer).CreateNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Notify_CreateNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServer).CreateNotify(ctx, req.(*CreateNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notify_UpdateNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServer).UpdateNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Notify_UpdateNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServer).UpdateNotify(ctx, req.(*UpdateNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notify_DeleteNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServer).DeleteNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Notify_DeleteNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServer).DeleteNotify(ctx, req.(*DeleteNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notify_GetNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServer).GetNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Notify_GetNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServer).GetNotify(ctx, req.(*GetNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notify_ListNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServer).ListNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Notify_ListNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServer).ListNotify(ctx, req.(*ListNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notify_SelectNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SelectNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotifyServer).SelectNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Notify_SelectNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotifyServer).SelectNotify(ctx, req.(*SelectNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Notify_ServiceDesc is the grpc.ServiceDesc for Notify service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Notify_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.server.prom.notify.Notify",
	HandlerType: (*NotifyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateNotify",
			Handler:    _Notify_CreateNotify_Handler,
		},
		{
			MethodName: "UpdateNotify",
			Handler:    _Notify_UpdateNotify_Handler,
		},
		{
			MethodName: "DeleteNotify",
			Handler:    _Notify_DeleteNotify_Handler,
		},
		{
			MethodName: "GetNotify",
			Handler:    _Notify_GetNotify_Handler,
		},
		{
			MethodName: "ListNotify",
			Handler:    _Notify_ListNotify_Handler,
		},
		{
			MethodName: "SelectNotify",
			Handler:    _Notify_SelectNotify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server/prom/notify/notify.proto",
}
