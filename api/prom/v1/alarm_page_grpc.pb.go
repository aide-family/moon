// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: prom/v1/alarm_page.proto

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

// AlarmPageClient is the client API for AlarmPage service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AlarmPageClient interface {
	// CreateAlarmPage creates a new alarm page.
	CreateAlarmPage(ctx context.Context, in *CreateAlarmPageRequest, opts ...grpc.CallOption) (*CreateAlarmPageReply, error)
	// UpdateAlarmPage updates an existing alarm page by id.
	UpdateAlarmPage(ctx context.Context, in *UpdateAlarmPageRequest, opts ...grpc.CallOption) (*UpdateAlarmPageReply, error)
	// UpdateAlarmPagesStatus updates an existing alarm page status by ids.
	UpdateAlarmPagesStatus(ctx context.Context, in *UpdateAlarmPagesStatusRequest, opts ...grpc.CallOption) (*UpdateAlarmPagesStatusReply, error)
	// DeleteAlarmPage deletes an existing alarm page by id.
	DeleteAlarmPage(ctx context.Context, in *DeleteAlarmPageRequest, opts ...grpc.CallOption) (*DeleteAlarmPageReply, error)
	// GetAlarmPage gets an existing alarm page by id.
	GetAlarmPage(ctx context.Context, in *GetAlarmPageRequest, opts ...grpc.CallOption) (*GetAlarmPageReply, error)
	// GetAlarmPage gets an existing alarm page by query and alarm page.
	ListAlarmPage(ctx context.Context, in *ListAlarmPageRequest, opts ...grpc.CallOption) (*ListAlarmPageReply, error)
}

type alarmPageClient struct {
	cc grpc.ClientConnInterface
}

func NewAlarmPageClient(cc grpc.ClientConnInterface) AlarmPageClient {
	return &alarmPageClient{cc}
}

func (c *alarmPageClient) CreateAlarmPage(ctx context.Context, in *CreateAlarmPageRequest, opts ...grpc.CallOption) (*CreateAlarmPageReply, error) {
	out := new(CreateAlarmPageReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.AlarmPage/CreateAlarmPage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *alarmPageClient) UpdateAlarmPage(ctx context.Context, in *UpdateAlarmPageRequest, opts ...grpc.CallOption) (*UpdateAlarmPageReply, error) {
	out := new(UpdateAlarmPageReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.AlarmPage/UpdateAlarmPage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *alarmPageClient) UpdateAlarmPagesStatus(ctx context.Context, in *UpdateAlarmPagesStatusRequest, opts ...grpc.CallOption) (*UpdateAlarmPagesStatusReply, error) {
	out := new(UpdateAlarmPagesStatusReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.AlarmPage/UpdateAlarmPagesStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *alarmPageClient) DeleteAlarmPage(ctx context.Context, in *DeleteAlarmPageRequest, opts ...grpc.CallOption) (*DeleteAlarmPageReply, error) {
	out := new(DeleteAlarmPageReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.AlarmPage/DeleteAlarmPage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *alarmPageClient) GetAlarmPage(ctx context.Context, in *GetAlarmPageRequest, opts ...grpc.CallOption) (*GetAlarmPageReply, error) {
	out := new(GetAlarmPageReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.AlarmPage/GetAlarmPage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *alarmPageClient) ListAlarmPage(ctx context.Context, in *ListAlarmPageRequest, opts ...grpc.CallOption) (*ListAlarmPageReply, error) {
	out := new(ListAlarmPageReply)
	err := c.cc.Invoke(ctx, "/api.prom.v1.AlarmPage/ListAlarmPage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AlarmPageServer is the server API for AlarmPage service.
// All implementations must embed UnimplementedAlarmPageServer
// for forward compatibility
type AlarmPageServer interface {
	// CreateAlarmPage creates a new alarm page.
	CreateAlarmPage(context.Context, *CreateAlarmPageRequest) (*CreateAlarmPageReply, error)
	// UpdateAlarmPage updates an existing alarm page by id.
	UpdateAlarmPage(context.Context, *UpdateAlarmPageRequest) (*UpdateAlarmPageReply, error)
	// UpdateAlarmPagesStatus updates an existing alarm page status by ids.
	UpdateAlarmPagesStatus(context.Context, *UpdateAlarmPagesStatusRequest) (*UpdateAlarmPagesStatusReply, error)
	// DeleteAlarmPage deletes an existing alarm page by id.
	DeleteAlarmPage(context.Context, *DeleteAlarmPageRequest) (*DeleteAlarmPageReply, error)
	// GetAlarmPage gets an existing alarm page by id.
	GetAlarmPage(context.Context, *GetAlarmPageRequest) (*GetAlarmPageReply, error)
	// GetAlarmPage gets an existing alarm page by query and alarm page.
	ListAlarmPage(context.Context, *ListAlarmPageRequest) (*ListAlarmPageReply, error)
	mustEmbedUnimplementedAlarmPageServer()
}

// UnimplementedAlarmPageServer must be embedded to have forward compatible implementations.
type UnimplementedAlarmPageServer struct {
}

func (UnimplementedAlarmPageServer) CreateAlarmPage(context.Context, *CreateAlarmPageRequest) (*CreateAlarmPageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAlarmPage not implemented")
}
func (UnimplementedAlarmPageServer) UpdateAlarmPage(context.Context, *UpdateAlarmPageRequest) (*UpdateAlarmPageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAlarmPage not implemented")
}
func (UnimplementedAlarmPageServer) UpdateAlarmPagesStatus(context.Context, *UpdateAlarmPagesStatusRequest) (*UpdateAlarmPagesStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAlarmPagesStatus not implemented")
}
func (UnimplementedAlarmPageServer) DeleteAlarmPage(context.Context, *DeleteAlarmPageRequest) (*DeleteAlarmPageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAlarmPage not implemented")
}
func (UnimplementedAlarmPageServer) GetAlarmPage(context.Context, *GetAlarmPageRequest) (*GetAlarmPageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAlarmPage not implemented")
}
func (UnimplementedAlarmPageServer) ListAlarmPage(context.Context, *ListAlarmPageRequest) (*ListAlarmPageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAlarmPage not implemented")
}
func (UnimplementedAlarmPageServer) mustEmbedUnimplementedAlarmPageServer() {}

// UnsafeAlarmPageServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AlarmPageServer will
// result in compilation errors.
type UnsafeAlarmPageServer interface {
	mustEmbedUnimplementedAlarmPageServer()
}

func RegisterAlarmPageServer(s grpc.ServiceRegistrar, srv AlarmPageServer) {
	s.RegisterService(&AlarmPage_ServiceDesc, srv)
}

func _AlarmPage_CreateAlarmPage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAlarmPageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlarmPageServer).CreateAlarmPage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.AlarmPage/CreateAlarmPage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlarmPageServer).CreateAlarmPage(ctx, req.(*CreateAlarmPageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AlarmPage_UpdateAlarmPage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAlarmPageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlarmPageServer).UpdateAlarmPage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.AlarmPage/UpdateAlarmPage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlarmPageServer).UpdateAlarmPage(ctx, req.(*UpdateAlarmPageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AlarmPage_UpdateAlarmPagesStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAlarmPagesStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlarmPageServer).UpdateAlarmPagesStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.AlarmPage/UpdateAlarmPagesStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlarmPageServer).UpdateAlarmPagesStatus(ctx, req.(*UpdateAlarmPagesStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AlarmPage_DeleteAlarmPage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAlarmPageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlarmPageServer).DeleteAlarmPage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.AlarmPage/DeleteAlarmPage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlarmPageServer).DeleteAlarmPage(ctx, req.(*DeleteAlarmPageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AlarmPage_GetAlarmPage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAlarmPageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlarmPageServer).GetAlarmPage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.AlarmPage/GetAlarmPage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlarmPageServer).GetAlarmPage(ctx, req.(*GetAlarmPageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AlarmPage_ListAlarmPage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAlarmPageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlarmPageServer).ListAlarmPage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.prom.v1.AlarmPage/ListAlarmPage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlarmPageServer).ListAlarmPage(ctx, req.(*ListAlarmPageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AlarmPage_ServiceDesc is the grpc.ServiceDesc for AlarmPage service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AlarmPage_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.prom.v1.AlarmPage",
	HandlerType: (*AlarmPageServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAlarmPage",
			Handler:    _AlarmPage_CreateAlarmPage_Handler,
		},
		{
			MethodName: "UpdateAlarmPage",
			Handler:    _AlarmPage_UpdateAlarmPage_Handler,
		},
		{
			MethodName: "UpdateAlarmPagesStatus",
			Handler:    _AlarmPage_UpdateAlarmPagesStatus_Handler,
		},
		{
			MethodName: "DeleteAlarmPage",
			Handler:    _AlarmPage_DeleteAlarmPage_Handler,
		},
		{
			MethodName: "GetAlarmPage",
			Handler:    _AlarmPage_GetAlarmPage_Handler,
		},
		{
			MethodName: "ListAlarmPage",
			Handler:    _AlarmPage_ListAlarmPage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "prom/v1/alarm_page.proto",
}
