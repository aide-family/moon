// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.4
// source: server/prom/strategy/strategy.proto

package strategy

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
	Strategy_CreateStrategy_FullMethodName            = "/api.server.prom.strategy.Strategy/CreateStrategy"
	Strategy_UpdateStrategy_FullMethodName            = "/api.server.prom.strategy.Strategy/UpdateStrategy"
	Strategy_BatchUpdateStrategyStatus_FullMethodName = "/api.server.prom.strategy.Strategy/BatchUpdateStrategyStatus"
	Strategy_DeleteStrategy_FullMethodName            = "/api.server.prom.strategy.Strategy/DeleteStrategy"
	Strategy_BatchDeleteStrategy_FullMethodName       = "/api.server.prom.strategy.Strategy/BatchDeleteStrategy"
	Strategy_GetStrategy_FullMethodName               = "/api.server.prom.strategy.Strategy/GetStrategy"
	Strategy_ListStrategy_FullMethodName              = "/api.server.prom.strategy.Strategy/ListStrategy"
	Strategy_SelectStrategy_FullMethodName            = "/api.server.prom.strategy.Strategy/SelectStrategy"
	Strategy_ExportStrategy_FullMethodName            = "/api.server.prom.strategy.Strategy/ExportStrategy"
	Strategy_GetStrategyNotifyObject_FullMethodName   = "/api.server.prom.strategy.Strategy/GetStrategyNotifyObject"
	Strategy_BindStrategyNotifyObject_FullMethodName  = "/api.server.prom.strategy.Strategy/BindStrategyNotifyObject"
	Strategy_TestNotifyTemplate_FullMethodName        = "/api.server.prom.strategy.Strategy/TestNotifyTemplate"
)

// StrategyClient is the client API for Strategy service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StrategyClient interface {
	// 创建策略
	CreateStrategy(ctx context.Context, in *CreateStrategyRequest, opts ...grpc.CallOption) (*CreateStrategyReply, error)
	// 更新策略
	UpdateStrategy(ctx context.Context, in *UpdateStrategyRequest, opts ...grpc.CallOption) (*UpdateStrategyReply, error)
	// 批量更新策略状态
	BatchUpdateStrategyStatus(ctx context.Context, in *BatchUpdateStrategyStatusRequest, opts ...grpc.CallOption) (*BatchUpdateStrategyStatusReply, error)
	// 删除策略
	DeleteStrategy(ctx context.Context, in *DeleteStrategyRequest, opts ...grpc.CallOption) (*DeleteStrategyReply, error)
	// 批量删除策略
	BatchDeleteStrategy(ctx context.Context, in *BatchDeleteStrategyRequest, opts ...grpc.CallOption) (*BatchDeleteStrategyReply, error)
	// 获取策略
	GetStrategy(ctx context.Context, in *GetStrategyRequest, opts ...grpc.CallOption) (*GetStrategyReply, error)
	// 获取策略列表
	ListStrategy(ctx context.Context, in *ListStrategyRequest, opts ...grpc.CallOption) (*ListStrategyReply, error)
	// 获取策略下拉列表
	SelectStrategy(ctx context.Context, in *SelectStrategyRequest, opts ...grpc.CallOption) (*SelectStrategyReply, error)
	// ExportStrategy 导出策略
	ExportStrategy(ctx context.Context, in *ExportStrategyRequest, opts ...grpc.CallOption) (*ExportStrategyReply, error)
	// 获取策略通知对象明细
	GetStrategyNotifyObject(ctx context.Context, in *GetStrategyNotifyObjectRequest, opts ...grpc.CallOption) (*GetStrategyNotifyObjectReply, error)
	// 绑定通知对象
	BindStrategyNotifyObject(ctx context.Context, in *BindStrategyNotifyObjectRequest, opts ...grpc.CallOption) (*BindStrategyNotifyObjectReply, error)
	// 测试hook模板
	TestNotifyTemplate(ctx context.Context, in *TestTemplateRequest, opts ...grpc.CallOption) (*TestTemplateReply, error)
}

type strategyClient struct {
	cc grpc.ClientConnInterface
}

func NewStrategyClient(cc grpc.ClientConnInterface) StrategyClient {
	return &strategyClient{cc}
}

func (c *strategyClient) CreateStrategy(ctx context.Context, in *CreateStrategyRequest, opts ...grpc.CallOption) (*CreateStrategyReply, error) {
	out := new(CreateStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_CreateStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) UpdateStrategy(ctx context.Context, in *UpdateStrategyRequest, opts ...grpc.CallOption) (*UpdateStrategyReply, error) {
	out := new(UpdateStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_UpdateStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) BatchUpdateStrategyStatus(ctx context.Context, in *BatchUpdateStrategyStatusRequest, opts ...grpc.CallOption) (*BatchUpdateStrategyStatusReply, error) {
	out := new(BatchUpdateStrategyStatusReply)
	err := c.cc.Invoke(ctx, Strategy_BatchUpdateStrategyStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) DeleteStrategy(ctx context.Context, in *DeleteStrategyRequest, opts ...grpc.CallOption) (*DeleteStrategyReply, error) {
	out := new(DeleteStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_DeleteStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) BatchDeleteStrategy(ctx context.Context, in *BatchDeleteStrategyRequest, opts ...grpc.CallOption) (*BatchDeleteStrategyReply, error) {
	out := new(BatchDeleteStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_BatchDeleteStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) GetStrategy(ctx context.Context, in *GetStrategyRequest, opts ...grpc.CallOption) (*GetStrategyReply, error) {
	out := new(GetStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_GetStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) ListStrategy(ctx context.Context, in *ListStrategyRequest, opts ...grpc.CallOption) (*ListStrategyReply, error) {
	out := new(ListStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_ListStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) SelectStrategy(ctx context.Context, in *SelectStrategyRequest, opts ...grpc.CallOption) (*SelectStrategyReply, error) {
	out := new(SelectStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_SelectStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) ExportStrategy(ctx context.Context, in *ExportStrategyRequest, opts ...grpc.CallOption) (*ExportStrategyReply, error) {
	out := new(ExportStrategyReply)
	err := c.cc.Invoke(ctx, Strategy_ExportStrategy_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) GetStrategyNotifyObject(ctx context.Context, in *GetStrategyNotifyObjectRequest, opts ...grpc.CallOption) (*GetStrategyNotifyObjectReply, error) {
	out := new(GetStrategyNotifyObjectReply)
	err := c.cc.Invoke(ctx, Strategy_GetStrategyNotifyObject_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) BindStrategyNotifyObject(ctx context.Context, in *BindStrategyNotifyObjectRequest, opts ...grpc.CallOption) (*BindStrategyNotifyObjectReply, error) {
	out := new(BindStrategyNotifyObjectReply)
	err := c.cc.Invoke(ctx, Strategy_BindStrategyNotifyObject_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyClient) TestNotifyTemplate(ctx context.Context, in *TestTemplateRequest, opts ...grpc.CallOption) (*TestTemplateReply, error) {
	out := new(TestTemplateReply)
	err := c.cc.Invoke(ctx, Strategy_TestNotifyTemplate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StrategyServer is the server API for Strategy service.
// All implementations must embed UnimplementedStrategyServer
// for forward compatibility
type StrategyServer interface {
	// 创建策略
	CreateStrategy(context.Context, *CreateStrategyRequest) (*CreateStrategyReply, error)
	// 更新策略
	UpdateStrategy(context.Context, *UpdateStrategyRequest) (*UpdateStrategyReply, error)
	// 批量更新策略状态
	BatchUpdateStrategyStatus(context.Context, *BatchUpdateStrategyStatusRequest) (*BatchUpdateStrategyStatusReply, error)
	// 删除策略
	DeleteStrategy(context.Context, *DeleteStrategyRequest) (*DeleteStrategyReply, error)
	// 批量删除策略
	BatchDeleteStrategy(context.Context, *BatchDeleteStrategyRequest) (*BatchDeleteStrategyReply, error)
	// 获取策略
	GetStrategy(context.Context, *GetStrategyRequest) (*GetStrategyReply, error)
	// 获取策略列表
	ListStrategy(context.Context, *ListStrategyRequest) (*ListStrategyReply, error)
	// 获取策略下拉列表
	SelectStrategy(context.Context, *SelectStrategyRequest) (*SelectStrategyReply, error)
	// ExportStrategy 导出策略
	ExportStrategy(context.Context, *ExportStrategyRequest) (*ExportStrategyReply, error)
	// 获取策略通知对象明细
	GetStrategyNotifyObject(context.Context, *GetStrategyNotifyObjectRequest) (*GetStrategyNotifyObjectReply, error)
	// 绑定通知对象
	BindStrategyNotifyObject(context.Context, *BindStrategyNotifyObjectRequest) (*BindStrategyNotifyObjectReply, error)
	// 测试hook模板
	TestNotifyTemplate(context.Context, *TestTemplateRequest) (*TestTemplateReply, error)
	mustEmbedUnimplementedStrategyServer()
}

// UnimplementedStrategyServer must be embedded to have forward compatible implementations.
type UnimplementedStrategyServer struct {
}

func (UnimplementedStrategyServer) CreateStrategy(context.Context, *CreateStrategyRequest) (*CreateStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateStrategy not implemented")
}
func (UnimplementedStrategyServer) UpdateStrategy(context.Context, *UpdateStrategyRequest) (*UpdateStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStrategy not implemented")
}
func (UnimplementedStrategyServer) BatchUpdateStrategyStatus(context.Context, *BatchUpdateStrategyStatusRequest) (*BatchUpdateStrategyStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchUpdateStrategyStatus not implemented")
}
func (UnimplementedStrategyServer) DeleteStrategy(context.Context, *DeleteStrategyRequest) (*DeleteStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteStrategy not implemented")
}
func (UnimplementedStrategyServer) BatchDeleteStrategy(context.Context, *BatchDeleteStrategyRequest) (*BatchDeleteStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchDeleteStrategy not implemented")
}
func (UnimplementedStrategyServer) GetStrategy(context.Context, *GetStrategyRequest) (*GetStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStrategy not implemented")
}
func (UnimplementedStrategyServer) ListStrategy(context.Context, *ListStrategyRequest) (*ListStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListStrategy not implemented")
}
func (UnimplementedStrategyServer) SelectStrategy(context.Context, *SelectStrategyRequest) (*SelectStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SelectStrategy not implemented")
}
func (UnimplementedStrategyServer) ExportStrategy(context.Context, *ExportStrategyRequest) (*ExportStrategyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportStrategy not implemented")
}
func (UnimplementedStrategyServer) GetStrategyNotifyObject(context.Context, *GetStrategyNotifyObjectRequest) (*GetStrategyNotifyObjectReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStrategyNotifyObject not implemented")
}
func (UnimplementedStrategyServer) BindStrategyNotifyObject(context.Context, *BindStrategyNotifyObjectRequest) (*BindStrategyNotifyObjectReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BindStrategyNotifyObject not implemented")
}
func (UnimplementedStrategyServer) TestNotifyTemplate(context.Context, *TestTemplateRequest) (*TestTemplateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestNotifyTemplate not implemented")
}
func (UnimplementedStrategyServer) mustEmbedUnimplementedStrategyServer() {}

// UnsafeStrategyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StrategyServer will
// result in compilation errors.
type UnsafeStrategyServer interface {
	mustEmbedUnimplementedStrategyServer()
}

func RegisterStrategyServer(s grpc.ServiceRegistrar, srv StrategyServer) {
	s.RegisterService(&Strategy_ServiceDesc, srv)
}

func _Strategy_CreateStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).CreateStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_CreateStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).CreateStrategy(ctx, req.(*CreateStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_UpdateStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).UpdateStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_UpdateStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).UpdateStrategy(ctx, req.(*UpdateStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_BatchUpdateStrategyStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchUpdateStrategyStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).BatchUpdateStrategyStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_BatchUpdateStrategyStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).BatchUpdateStrategyStatus(ctx, req.(*BatchUpdateStrategyStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_DeleteStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).DeleteStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_DeleteStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).DeleteStrategy(ctx, req.(*DeleteStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_BatchDeleteStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchDeleteStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).BatchDeleteStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_BatchDeleteStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).BatchDeleteStrategy(ctx, req.(*BatchDeleteStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_GetStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).GetStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_GetStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).GetStrategy(ctx, req.(*GetStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_ListStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).ListStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_ListStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).ListStrategy(ctx, req.(*ListStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_SelectStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SelectStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).SelectStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_SelectStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).SelectStrategy(ctx, req.(*SelectStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_ExportStrategy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportStrategyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).ExportStrategy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_ExportStrategy_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).ExportStrategy(ctx, req.(*ExportStrategyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_GetStrategyNotifyObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStrategyNotifyObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).GetStrategyNotifyObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_GetStrategyNotifyObject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).GetStrategyNotifyObject(ctx, req.(*GetStrategyNotifyObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_BindStrategyNotifyObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BindStrategyNotifyObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).BindStrategyNotifyObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_BindStrategyNotifyObject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).BindStrategyNotifyObject(ctx, req.(*BindStrategyNotifyObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Strategy_TestNotifyTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyServer).TestNotifyTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Strategy_TestNotifyTemplate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyServer).TestNotifyTemplate(ctx, req.(*TestTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Strategy_ServiceDesc is the grpc.ServiceDesc for Strategy service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Strategy_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.server.prom.strategy.Strategy",
	HandlerType: (*StrategyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateStrategy",
			Handler:    _Strategy_CreateStrategy_Handler,
		},
		{
			MethodName: "UpdateStrategy",
			Handler:    _Strategy_UpdateStrategy_Handler,
		},
		{
			MethodName: "BatchUpdateStrategyStatus",
			Handler:    _Strategy_BatchUpdateStrategyStatus_Handler,
		},
		{
			MethodName: "DeleteStrategy",
			Handler:    _Strategy_DeleteStrategy_Handler,
		},
		{
			MethodName: "BatchDeleteStrategy",
			Handler:    _Strategy_BatchDeleteStrategy_Handler,
		},
		{
			MethodName: "GetStrategy",
			Handler:    _Strategy_GetStrategy_Handler,
		},
		{
			MethodName: "ListStrategy",
			Handler:    _Strategy_ListStrategy_Handler,
		},
		{
			MethodName: "SelectStrategy",
			Handler:    _Strategy_SelectStrategy_Handler,
		},
		{
			MethodName: "ExportStrategy",
			Handler:    _Strategy_ExportStrategy_Handler,
		},
		{
			MethodName: "GetStrategyNotifyObject",
			Handler:    _Strategy_GetStrategyNotifyObject_Handler,
		},
		{
			MethodName: "BindStrategyNotifyObject",
			Handler:    _Strategy_BindStrategyNotifyObject_Handler,
		},
		{
			MethodName: "TestNotifyTemplate",
			Handler:    _Strategy_TestNotifyTemplate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server/prom/strategy/strategy.proto",
}
