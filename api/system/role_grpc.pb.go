// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: system/role.proto

package system

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

// RoleClient is the client API for Role service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RoleClient interface {
	// 创建角色
	CreateRole(ctx context.Context, in *CreateRoleRequest, opts ...grpc.CallOption) (*CreateRoleReply, error)
	// 更新角色
	UpdateRole(ctx context.Context, in *UpdateRoleRequest, opts ...grpc.CallOption) (*UpdateRoleReply, error)
	// 删除角色
	DeleteRole(ctx context.Context, in *DeleteRoleRequest, opts ...grpc.CallOption) (*DeleteRoleReply, error)
	// 获取角色
	GetRole(ctx context.Context, in *GetRoleRequest, opts ...grpc.CallOption) (*GetRoleReply, error)
	// 获取角色列表
	ListRole(ctx context.Context, in *ListRoleRequest, opts ...grpc.CallOption) (*ListRoleReply, error)
	// 获取角色下拉列表
	SelectRole(ctx context.Context, in *SelectRoleRequest, opts ...grpc.CallOption) (*SelectRoleReply, error)
	// 关联角色接口
	RelateApi(ctx context.Context, in *RelateApiRequest, opts ...grpc.CallOption) (*RelateApiReply, error)
	// 修改角色状态
	EditRoleStatus(ctx context.Context, in *EditRoleStatusRequest, opts ...grpc.CallOption) (*EditRoleStatusReply, error)
}

type roleClient struct {
	cc grpc.ClientConnInterface
}

func NewRoleClient(cc grpc.ClientConnInterface) RoleClient {
	return &roleClient{cc}
}

func (c *roleClient) CreateRole(ctx context.Context, in *CreateRoleRequest, opts ...grpc.CallOption) (*CreateRoleReply, error) {
	out := new(CreateRoleReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/CreateRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) UpdateRole(ctx context.Context, in *UpdateRoleRequest, opts ...grpc.CallOption) (*UpdateRoleReply, error) {
	out := new(UpdateRoleReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/UpdateRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) DeleteRole(ctx context.Context, in *DeleteRoleRequest, opts ...grpc.CallOption) (*DeleteRoleReply, error) {
	out := new(DeleteRoleReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/DeleteRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) GetRole(ctx context.Context, in *GetRoleRequest, opts ...grpc.CallOption) (*GetRoleReply, error) {
	out := new(GetRoleReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/GetRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) ListRole(ctx context.Context, in *ListRoleRequest, opts ...grpc.CallOption) (*ListRoleReply, error) {
	out := new(ListRoleReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/ListRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) SelectRole(ctx context.Context, in *SelectRoleRequest, opts ...grpc.CallOption) (*SelectRoleReply, error) {
	out := new(SelectRoleReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/SelectRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) RelateApi(ctx context.Context, in *RelateApiRequest, opts ...grpc.CallOption) (*RelateApiReply, error) {
	out := new(RelateApiReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/RelateApi", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) EditRoleStatus(ctx context.Context, in *EditRoleStatusRequest, opts ...grpc.CallOption) (*EditRoleStatusReply, error) {
	out := new(EditRoleStatusReply)
	err := c.cc.Invoke(ctx, "/api.system.Role/EditRoleStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RoleServer is the server API for Role service.
// All implementations must embed UnimplementedRoleServer
// for forward compatibility
type RoleServer interface {
	// 创建角色
	CreateRole(context.Context, *CreateRoleRequest) (*CreateRoleReply, error)
	// 更新角色
	UpdateRole(context.Context, *UpdateRoleRequest) (*UpdateRoleReply, error)
	// 删除角色
	DeleteRole(context.Context, *DeleteRoleRequest) (*DeleteRoleReply, error)
	// 获取角色
	GetRole(context.Context, *GetRoleRequest) (*GetRoleReply, error)
	// 获取角色列表
	ListRole(context.Context, *ListRoleRequest) (*ListRoleReply, error)
	// 获取角色下拉列表
	SelectRole(context.Context, *SelectRoleRequest) (*SelectRoleReply, error)
	// 关联角色接口
	RelateApi(context.Context, *RelateApiRequest) (*RelateApiReply, error)
	// 修改角色状态
	EditRoleStatus(context.Context, *EditRoleStatusRequest) (*EditRoleStatusReply, error)
	mustEmbedUnimplementedRoleServer()
}

// UnimplementedRoleServer must be embedded to have forward compatible implementations.
type UnimplementedRoleServer struct {
}

func (UnimplementedRoleServer) CreateRole(context.Context, *CreateRoleRequest) (*CreateRoleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRole not implemented")
}
func (UnimplementedRoleServer) UpdateRole(context.Context, *UpdateRoleRequest) (*UpdateRoleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRole not implemented")
}
func (UnimplementedRoleServer) DeleteRole(context.Context, *DeleteRoleRequest) (*DeleteRoleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRole not implemented")
}
func (UnimplementedRoleServer) GetRole(context.Context, *GetRoleRequest) (*GetRoleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRole not implemented")
}
func (UnimplementedRoleServer) ListRole(context.Context, *ListRoleRequest) (*ListRoleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRole not implemented")
}
func (UnimplementedRoleServer) SelectRole(context.Context, *SelectRoleRequest) (*SelectRoleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SelectRole not implemented")
}
func (UnimplementedRoleServer) RelateApi(context.Context, *RelateApiRequest) (*RelateApiReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RelateApi not implemented")
}
func (UnimplementedRoleServer) EditRoleStatus(context.Context, *EditRoleStatusRequest) (*EditRoleStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EditRoleStatus not implemented")
}
func (UnimplementedRoleServer) mustEmbedUnimplementedRoleServer() {}

// UnsafeRoleServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RoleServer will
// result in compilation errors.
type UnsafeRoleServer interface {
	mustEmbedUnimplementedRoleServer()
}

func RegisterRoleServer(s grpc.ServiceRegistrar, srv RoleServer) {
	s.RegisterService(&Role_ServiceDesc, srv)
}

func _Role_CreateRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).CreateRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/CreateRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).CreateRole(ctx, req.(*CreateRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_UpdateRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).UpdateRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/UpdateRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).UpdateRole(ctx, req.(*UpdateRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_DeleteRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).DeleteRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/DeleteRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).DeleteRole(ctx, req.(*DeleteRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_GetRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).GetRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/GetRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).GetRole(ctx, req.(*GetRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_ListRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).ListRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/ListRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).ListRole(ctx, req.(*ListRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_SelectRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SelectRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).SelectRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/SelectRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).SelectRole(ctx, req.(*SelectRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_RelateApi_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RelateApiRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).RelateApi(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/RelateApi",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).RelateApi(ctx, req.(*RelateApiRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_EditRoleStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EditRoleStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).EditRoleStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.system.Role/EditRoleStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).EditRoleStatus(ctx, req.(*EditRoleStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Role_ServiceDesc is the grpc.ServiceDesc for Role service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Role_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.system.Role",
	HandlerType: (*RoleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateRole",
			Handler:    _Role_CreateRole_Handler,
		},
		{
			MethodName: "UpdateRole",
			Handler:    _Role_UpdateRole_Handler,
		},
		{
			MethodName: "DeleteRole",
			Handler:    _Role_DeleteRole_Handler,
		},
		{
			MethodName: "GetRole",
			Handler:    _Role_GetRole_Handler,
		},
		{
			MethodName: "ListRole",
			Handler:    _Role_ListRole_Handler,
		},
		{
			MethodName: "SelectRole",
			Handler:    _Role_SelectRole_Handler,
		},
		{
			MethodName: "RelateApi",
			Handler:    _Role_RelateApi_Handler,
		},
		{
			MethodName: "EditRoleStatus",
			Handler:    _Role_EditRoleStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "system/role.proto",
}
