// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.4
// source: server/prom/strategy/group/group.proto

package group

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
	Group_CreateGroup_FullMethodName            = "/api.server.prom.strategy.group.Group/CreateGroup"
	Group_UpdateGroup_FullMethodName            = "/api.server.prom.strategy.group.Group/UpdateGroup"
	Group_BatchUpdateGroupStatus_FullMethodName = "/api.server.prom.strategy.group.Group/BatchUpdateGroupStatus"
	Group_DeleteGroup_FullMethodName            = "/api.server.prom.strategy.group.Group/DeleteGroup"
	Group_BatchDeleteGroup_FullMethodName       = "/api.server.prom.strategy.group.Group/BatchDeleteGroup"
	Group_GetGroup_FullMethodName               = "/api.server.prom.strategy.group.Group/GetGroup"
	Group_ListGroup_FullMethodName              = "/api.server.prom.strategy.group.Group/ListGroup"
	Group_ListAllGroupDetail_FullMethodName     = "/api.server.prom.strategy.group.Group/ListAllGroupDetail"
	Group_ListAllGroupDetailV2_FullMethodName   = "/api.server.prom.strategy.group.Group/ListAllGroupDetailV2"
	Group_SelectGroup_FullMethodName            = "/api.server.prom.strategy.group.Group/SelectGroup"
	Group_ImportGroup_FullMethodName            = "/api.server.prom.strategy.group.Group/ImportGroup"
	Group_ExportGroup_FullMethodName            = "/api.server.prom.strategy.group.Group/ExportGroup"
)

// GroupClient is the client API for Group service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GroupClient interface {
	// 创建策略组
	CreateGroup(ctx context.Context, in *CreateGroupRequest, opts ...grpc.CallOption) (*CreateGroupReply, error)
	// 更新策略组
	UpdateGroup(ctx context.Context, in *UpdateGroupRequest, opts ...grpc.CallOption) (*UpdateGroupReply, error)
	// 批量更新策略组状态
	BatchUpdateGroupStatus(ctx context.Context, in *BatchUpdateGroupStatusRequest, opts ...grpc.CallOption) (*BatchUpdateGroupStatusReply, error)
	// 删除策略组
	DeleteGroup(ctx context.Context, in *DeleteGroupRequest, opts ...grpc.CallOption) (*DeleteGroupReply, error)
	// 批量删除策略组
	BatchDeleteGroup(ctx context.Context, in *BatchDeleteGroupRequest, opts ...grpc.CallOption) (*BatchDeleteGroupReply, error)
	// GetGroup 获取策略组
	GetGroup(ctx context.Context, in *GetGroupRequest, opts ...grpc.CallOption) (*GetGroupReply, error)
	// 获取策略组列表
	ListGroup(ctx context.Context, in *ListGroupRequest, opts ...grpc.CallOption) (*ListGroupReply, error)
	// 获取策略组列表明细
	ListAllGroupDetail(ctx context.Context, in *ListAllGroupDetailRequest, opts ...grpc.CallOption) (*ListAllGroupDetailReply, error)
	// 获取策略组列表明细V2
	ListAllGroupDetailV2(ctx context.Context, in *ListAllGroupDetailV2Request, opts ...grpc.CallOption) (*ListAllGroupDetailV2Reply, error)
	// 获取策略组下拉列表
	SelectGroup(ctx context.Context, in *SelectGroupRequest, opts ...grpc.CallOption) (*SelectGroupReply, error)
	// 导入策略组
	ImportGroup(ctx context.Context, in *ImportGroupRequest, opts ...grpc.CallOption) (*ImportGroupReply, error)
	// 导出策略组
	ExportGroup(ctx context.Context, in *ExportGroupRequest, opts ...grpc.CallOption) (*ExportGroupReply, error)
}

type groupClient struct {
	cc grpc.ClientConnInterface
}

func NewGroupClient(cc grpc.ClientConnInterface) GroupClient {
	return &groupClient{cc}
}

func (c *groupClient) CreateGroup(ctx context.Context, in *CreateGroupRequest, opts ...grpc.CallOption) (*CreateGroupReply, error) {
	out := new(CreateGroupReply)
	err := c.cc.Invoke(ctx, Group_CreateGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) UpdateGroup(ctx context.Context, in *UpdateGroupRequest, opts ...grpc.CallOption) (*UpdateGroupReply, error) {
	out := new(UpdateGroupReply)
	err := c.cc.Invoke(ctx, Group_UpdateGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) BatchUpdateGroupStatus(ctx context.Context, in *BatchUpdateGroupStatusRequest, opts ...grpc.CallOption) (*BatchUpdateGroupStatusReply, error) {
	out := new(BatchUpdateGroupStatusReply)
	err := c.cc.Invoke(ctx, Group_BatchUpdateGroupStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) DeleteGroup(ctx context.Context, in *DeleteGroupRequest, opts ...grpc.CallOption) (*DeleteGroupReply, error) {
	out := new(DeleteGroupReply)
	err := c.cc.Invoke(ctx, Group_DeleteGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) BatchDeleteGroup(ctx context.Context, in *BatchDeleteGroupRequest, opts ...grpc.CallOption) (*BatchDeleteGroupReply, error) {
	out := new(BatchDeleteGroupReply)
	err := c.cc.Invoke(ctx, Group_BatchDeleteGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) GetGroup(ctx context.Context, in *GetGroupRequest, opts ...grpc.CallOption) (*GetGroupReply, error) {
	out := new(GetGroupReply)
	err := c.cc.Invoke(ctx, Group_GetGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) ListGroup(ctx context.Context, in *ListGroupRequest, opts ...grpc.CallOption) (*ListGroupReply, error) {
	out := new(ListGroupReply)
	err := c.cc.Invoke(ctx, Group_ListGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) ListAllGroupDetail(ctx context.Context, in *ListAllGroupDetailRequest, opts ...grpc.CallOption) (*ListAllGroupDetailReply, error) {
	out := new(ListAllGroupDetailReply)
	err := c.cc.Invoke(ctx, Group_ListAllGroupDetail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) ListAllGroupDetailV2(ctx context.Context, in *ListAllGroupDetailV2Request, opts ...grpc.CallOption) (*ListAllGroupDetailV2Reply, error) {
	out := new(ListAllGroupDetailV2Reply)
	err := c.cc.Invoke(ctx, Group_ListAllGroupDetailV2_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) SelectGroup(ctx context.Context, in *SelectGroupRequest, opts ...grpc.CallOption) (*SelectGroupReply, error) {
	out := new(SelectGroupReply)
	err := c.cc.Invoke(ctx, Group_SelectGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) ImportGroup(ctx context.Context, in *ImportGroupRequest, opts ...grpc.CallOption) (*ImportGroupReply, error) {
	out := new(ImportGroupReply)
	err := c.cc.Invoke(ctx, Group_ImportGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupClient) ExportGroup(ctx context.Context, in *ExportGroupRequest, opts ...grpc.CallOption) (*ExportGroupReply, error) {
	out := new(ExportGroupReply)
	err := c.cc.Invoke(ctx, Group_ExportGroup_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GroupServer is the server API for Group service.
// All implementations must embed UnimplementedGroupServer
// for forward compatibility
type GroupServer interface {
	// 创建策略组
	CreateGroup(context.Context, *CreateGroupRequest) (*CreateGroupReply, error)
	// 更新策略组
	UpdateGroup(context.Context, *UpdateGroupRequest) (*UpdateGroupReply, error)
	// 批量更新策略组状态
	BatchUpdateGroupStatus(context.Context, *BatchUpdateGroupStatusRequest) (*BatchUpdateGroupStatusReply, error)
	// 删除策略组
	DeleteGroup(context.Context, *DeleteGroupRequest) (*DeleteGroupReply, error)
	// 批量删除策略组
	BatchDeleteGroup(context.Context, *BatchDeleteGroupRequest) (*BatchDeleteGroupReply, error)
	// GetGroup 获取策略组
	GetGroup(context.Context, *GetGroupRequest) (*GetGroupReply, error)
	// 获取策略组列表
	ListGroup(context.Context, *ListGroupRequest) (*ListGroupReply, error)
	// 获取策略组列表明细
	ListAllGroupDetail(context.Context, *ListAllGroupDetailRequest) (*ListAllGroupDetailReply, error)
	// 获取策略组列表明细V2
	ListAllGroupDetailV2(context.Context, *ListAllGroupDetailV2Request) (*ListAllGroupDetailV2Reply, error)
	// 获取策略组下拉列表
	SelectGroup(context.Context, *SelectGroupRequest) (*SelectGroupReply, error)
	// 导入策略组
	ImportGroup(context.Context, *ImportGroupRequest) (*ImportGroupReply, error)
	// 导出策略组
	ExportGroup(context.Context, *ExportGroupRequest) (*ExportGroupReply, error)
	mustEmbedUnimplementedGroupServer()
}

// UnimplementedGroupServer must be embedded to have forward compatible implementations.
type UnimplementedGroupServer struct {
}

func (UnimplementedGroupServer) CreateGroup(context.Context, *CreateGroupRequest) (*CreateGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroup not implemented")
}
func (UnimplementedGroupServer) UpdateGroup(context.Context, *UpdateGroupRequest) (*UpdateGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroup not implemented")
}
func (UnimplementedGroupServer) BatchUpdateGroupStatus(context.Context, *BatchUpdateGroupStatusRequest) (*BatchUpdateGroupStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchUpdateGroupStatus not implemented")
}
func (UnimplementedGroupServer) DeleteGroup(context.Context, *DeleteGroupRequest) (*DeleteGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteGroup not implemented")
}
func (UnimplementedGroupServer) BatchDeleteGroup(context.Context, *BatchDeleteGroupRequest) (*BatchDeleteGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchDeleteGroup not implemented")
}
func (UnimplementedGroupServer) GetGroup(context.Context, *GetGroupRequest) (*GetGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroup not implemented")
}
func (UnimplementedGroupServer) ListGroup(context.Context, *ListGroupRequest) (*ListGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListGroup not implemented")
}
func (UnimplementedGroupServer) ListAllGroupDetail(context.Context, *ListAllGroupDetailRequest) (*ListAllGroupDetailReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAllGroupDetail not implemented")
}
func (UnimplementedGroupServer) ListAllGroupDetailV2(context.Context, *ListAllGroupDetailV2Request) (*ListAllGroupDetailV2Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAllGroupDetailV2 not implemented")
}
func (UnimplementedGroupServer) SelectGroup(context.Context, *SelectGroupRequest) (*SelectGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SelectGroup not implemented")
}
func (UnimplementedGroupServer) ImportGroup(context.Context, *ImportGroupRequest) (*ImportGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportGroup not implemented")
}
func (UnimplementedGroupServer) ExportGroup(context.Context, *ExportGroupRequest) (*ExportGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportGroup not implemented")
}
func (UnimplementedGroupServer) mustEmbedUnimplementedGroupServer() {}

// UnsafeGroupServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GroupServer will
// result in compilation errors.
type UnsafeGroupServer interface {
	mustEmbedUnimplementedGroupServer()
}

func RegisterGroupServer(s grpc.ServiceRegistrar, srv GroupServer) {
	s.RegisterService(&Group_ServiceDesc, srv)
}

func _Group_CreateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).CreateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_CreateGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).CreateGroup(ctx, req.(*CreateGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_UpdateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).UpdateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_UpdateGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).UpdateGroup(ctx, req.(*UpdateGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_BatchUpdateGroupStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchUpdateGroupStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).BatchUpdateGroupStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_BatchUpdateGroupStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).BatchUpdateGroupStatus(ctx, req.(*BatchUpdateGroupStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_DeleteGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).DeleteGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_DeleteGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).DeleteGroup(ctx, req.(*DeleteGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_BatchDeleteGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchDeleteGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).BatchDeleteGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_BatchDeleteGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).BatchDeleteGroup(ctx, req.(*BatchDeleteGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_GetGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).GetGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_GetGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).GetGroup(ctx, req.(*GetGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_ListGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).ListGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_ListGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).ListGroup(ctx, req.(*ListGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_ListAllGroupDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAllGroupDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).ListAllGroupDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_ListAllGroupDetail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).ListAllGroupDetail(ctx, req.(*ListAllGroupDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_ListAllGroupDetailV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAllGroupDetailV2Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).ListAllGroupDetailV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_ListAllGroupDetailV2_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).ListAllGroupDetailV2(ctx, req.(*ListAllGroupDetailV2Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_SelectGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SelectGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).SelectGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_SelectGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).SelectGroup(ctx, req.(*SelectGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_ImportGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).ImportGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_ImportGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).ImportGroup(ctx, req.(*ImportGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Group_ExportGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupServer).ExportGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Group_ExportGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupServer).ExportGroup(ctx, req.(*ExportGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Group_ServiceDesc is the grpc.ServiceDesc for Group service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Group_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.server.prom.strategy.group.Group",
	HandlerType: (*GroupServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateGroup",
			Handler:    _Group_CreateGroup_Handler,
		},
		{
			MethodName: "UpdateGroup",
			Handler:    _Group_UpdateGroup_Handler,
		},
		{
			MethodName: "BatchUpdateGroupStatus",
			Handler:    _Group_BatchUpdateGroupStatus_Handler,
		},
		{
			MethodName: "DeleteGroup",
			Handler:    _Group_DeleteGroup_Handler,
		},
		{
			MethodName: "BatchDeleteGroup",
			Handler:    _Group_BatchDeleteGroup_Handler,
		},
		{
			MethodName: "GetGroup",
			Handler:    _Group_GetGroup_Handler,
		},
		{
			MethodName: "ListGroup",
			Handler:    _Group_ListGroup_Handler,
		},
		{
			MethodName: "ListAllGroupDetail",
			Handler:    _Group_ListAllGroupDetail_Handler,
		},
		{
			MethodName: "ListAllGroupDetailV2",
			Handler:    _Group_ListAllGroupDetailV2_Handler,
		},
		{
			MethodName: "SelectGroup",
			Handler:    _Group_SelectGroup_Handler,
		},
		{
			MethodName: "ImportGroup",
			Handler:    _Group_ImportGroup_Handler,
		},
		{
			MethodName: "ExportGroup",
			Handler:    _Group_ExportGroup_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server/prom/strategy/group/group.proto",
}
