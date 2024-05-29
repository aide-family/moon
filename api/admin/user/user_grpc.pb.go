// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: admin/user/user.proto

package user

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

// UserClient is the client API for User service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserClient interface {
	// 创建用户
	CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserReply, error)
	// 更新用户
	UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserReply, error)
	// 删除用户
	DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserReply, error)
	// 获取用户
	GetUser(ctx context.Context, in *GetUserRequest, opts ...grpc.CallOption) (*GetUserReply, error)
	// 列表用户
	ListUser(ctx context.Context, in *ListUserRequest, opts ...grpc.CallOption) (*ListUserReply, error)
	// 批量修改用户状态
	BatchUpdateUserStatus(ctx context.Context, in *BatchUpdateUserStatusRequest, opts ...grpc.CallOption) (*BatchUpdateUserStatusReply, error)
	// 重置用户密码
	ResetUserPassword(ctx context.Context, in *ResetUserPasswordRequest, opts ...grpc.CallOption) (*ResetUserPasswordReply, error)
	// 用户修改密码
	ResetUserPasswordBySelf(ctx context.Context, in *ResetUserPasswordBySelfRequest, opts ...grpc.CallOption) (*ResetUserPasswordBySelfReply, error)
	// 获取用户下拉列表
	GetUserSelectList(ctx context.Context, in *GetUserSelectListRequest, opts ...grpc.CallOption) (*GetUserSelectListReply, error)
	// 修改电话号码
	UpdateUserPhone(ctx context.Context, in *UpdateUserPhoneRequest, opts ...grpc.CallOption) (*UpdateUserPhoneReply, error)
	// 修改邮箱
	UpdateUserEmail(ctx context.Context, in *UpdateUserEmailRequest, opts ...grpc.CallOption) (*UpdateUserEmailReply, error)
	// 修改用户头像
	UpdateUserAvatar(ctx context.Context, in *UpdateUserAvatarRequest, opts ...grpc.CallOption) (*UpdateUserAvatarReply, error)
}

type userClient struct {
	cc grpc.ClientConnInterface
}

func NewUserClient(cc grpc.ClientConnInterface) UserClient {
	return &userClient{cc}
}

func (c *userClient) CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserReply, error) {
	out := new(CreateUserReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/CreateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) UpdateUser(ctx context.Context, in *UpdateUserRequest, opts ...grpc.CallOption) (*UpdateUserReply, error) {
	out := new(UpdateUserReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/UpdateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserReply, error) {
	out := new(DeleteUserReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/DeleteUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetUser(ctx context.Context, in *GetUserRequest, opts ...grpc.CallOption) (*GetUserReply, error) {
	out := new(GetUserReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/GetUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) ListUser(ctx context.Context, in *ListUserRequest, opts ...grpc.CallOption) (*ListUserReply, error) {
	out := new(ListUserReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/ListUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) BatchUpdateUserStatus(ctx context.Context, in *BatchUpdateUserStatusRequest, opts ...grpc.CallOption) (*BatchUpdateUserStatusReply, error) {
	out := new(BatchUpdateUserStatusReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/BatchUpdateUserStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) ResetUserPassword(ctx context.Context, in *ResetUserPasswordRequest, opts ...grpc.CallOption) (*ResetUserPasswordReply, error) {
	out := new(ResetUserPasswordReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/ResetUserPassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) ResetUserPasswordBySelf(ctx context.Context, in *ResetUserPasswordBySelfRequest, opts ...grpc.CallOption) (*ResetUserPasswordBySelfReply, error) {
	out := new(ResetUserPasswordBySelfReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/ResetUserPasswordBySelf", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) GetUserSelectList(ctx context.Context, in *GetUserSelectListRequest, opts ...grpc.CallOption) (*GetUserSelectListReply, error) {
	out := new(GetUserSelectListReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/GetUserSelectList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) UpdateUserPhone(ctx context.Context, in *UpdateUserPhoneRequest, opts ...grpc.CallOption) (*UpdateUserPhoneReply, error) {
	out := new(UpdateUserPhoneReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/UpdateUserPhone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) UpdateUserEmail(ctx context.Context, in *UpdateUserEmailRequest, opts ...grpc.CallOption) (*UpdateUserEmailReply, error) {
	out := new(UpdateUserEmailReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/UpdateUserEmail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userClient) UpdateUserAvatar(ctx context.Context, in *UpdateUserAvatarRequest, opts ...grpc.CallOption) (*UpdateUserAvatarReply, error) {
	out := new(UpdateUserAvatarReply)
	err := c.cc.Invoke(ctx, "/api.admin.user.User/UpdateUserAvatar", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServer is the server API for User service.
// All implementations must embed UnimplementedUserServer
// for forward compatibility
type UserServer interface {
	// 创建用户
	CreateUser(context.Context, *CreateUserRequest) (*CreateUserReply, error)
	// 更新用户
	UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserReply, error)
	// 删除用户
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserReply, error)
	// 获取用户
	GetUser(context.Context, *GetUserRequest) (*GetUserReply, error)
	// 列表用户
	ListUser(context.Context, *ListUserRequest) (*ListUserReply, error)
	// 批量修改用户状态
	BatchUpdateUserStatus(context.Context, *BatchUpdateUserStatusRequest) (*BatchUpdateUserStatusReply, error)
	// 重置用户密码
	ResetUserPassword(context.Context, *ResetUserPasswordRequest) (*ResetUserPasswordReply, error)
	// 用户修改密码
	ResetUserPasswordBySelf(context.Context, *ResetUserPasswordBySelfRequest) (*ResetUserPasswordBySelfReply, error)
	// 获取用户下拉列表
	GetUserSelectList(context.Context, *GetUserSelectListRequest) (*GetUserSelectListReply, error)
	// 修改电话号码
	UpdateUserPhone(context.Context, *UpdateUserPhoneRequest) (*UpdateUserPhoneReply, error)
	// 修改邮箱
	UpdateUserEmail(context.Context, *UpdateUserEmailRequest) (*UpdateUserEmailReply, error)
	// 修改用户头像
	UpdateUserAvatar(context.Context, *UpdateUserAvatarRequest) (*UpdateUserAvatarReply, error)
	mustEmbedUnimplementedUserServer()
}

// UnimplementedUserServer must be embedded to have forward compatible implementations.
type UnimplementedUserServer struct {
}

func (UnimplementedUserServer) CreateUser(context.Context, *CreateUserRequest) (*CreateUserReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}
func (UnimplementedUserServer) UpdateUser(context.Context, *UpdateUserRequest) (*UpdateUserReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
func (UnimplementedUserServer) DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedUserServer) GetUser(context.Context, *GetUserRequest) (*GetUserReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedUserServer) ListUser(context.Context, *ListUserRequest) (*ListUserReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUser not implemented")
}
func (UnimplementedUserServer) BatchUpdateUserStatus(context.Context, *BatchUpdateUserStatusRequest) (*BatchUpdateUserStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchUpdateUserStatus not implemented")
}
func (UnimplementedUserServer) ResetUserPassword(context.Context, *ResetUserPasswordRequest) (*ResetUserPasswordReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetUserPassword not implemented")
}
func (UnimplementedUserServer) ResetUserPasswordBySelf(context.Context, *ResetUserPasswordBySelfRequest) (*ResetUserPasswordBySelfReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetUserPasswordBySelf not implemented")
}
func (UnimplementedUserServer) GetUserSelectList(context.Context, *GetUserSelectListRequest) (*GetUserSelectListReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserSelectList not implemented")
}
func (UnimplementedUserServer) UpdateUserPhone(context.Context, *UpdateUserPhoneRequest) (*UpdateUserPhoneReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserPhone not implemented")
}
func (UnimplementedUserServer) UpdateUserEmail(context.Context, *UpdateUserEmailRequest) (*UpdateUserEmailReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserEmail not implemented")
}
func (UnimplementedUserServer) UpdateUserAvatar(context.Context, *UpdateUserAvatarRequest) (*UpdateUserAvatarReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserAvatar not implemented")
}
func (UnimplementedUserServer) mustEmbedUnimplementedUserServer() {}

// UnsafeUserServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServer will
// result in compilation errors.
type UnsafeUserServer interface {
	mustEmbedUnimplementedUserServer()
}

func RegisterUserServer(s grpc.ServiceRegistrar, srv UserServer) {
	s.RegisterService(&User_ServiceDesc, srv)
}

func _User_CreateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).CreateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/CreateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).CreateUser(ctx, req.(*CreateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_UpdateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).UpdateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/UpdateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).UpdateUser(ctx, req.(*UpdateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/DeleteUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).DeleteUser(ctx, req.(*DeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/GetUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetUser(ctx, req.(*GetUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_ListUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).ListUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/ListUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).ListUser(ctx, req.(*ListUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_BatchUpdateUserStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchUpdateUserStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).BatchUpdateUserStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/BatchUpdateUserStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).BatchUpdateUserStatus(ctx, req.(*BatchUpdateUserStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_ResetUserPassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResetUserPasswordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).ResetUserPassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/ResetUserPassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).ResetUserPassword(ctx, req.(*ResetUserPasswordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_ResetUserPasswordBySelf_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResetUserPasswordBySelfRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).ResetUserPasswordBySelf(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/ResetUserPasswordBySelf",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).ResetUserPasswordBySelf(ctx, req.(*ResetUserPasswordBySelfRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_GetUserSelectList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserSelectListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).GetUserSelectList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/GetUserSelectList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).GetUserSelectList(ctx, req.(*GetUserSelectListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_UpdateUserPhone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserPhoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).UpdateUserPhone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/UpdateUserPhone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).UpdateUserPhone(ctx, req.(*UpdateUserPhoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_UpdateUserEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).UpdateUserEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/UpdateUserEmail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).UpdateUserEmail(ctx, req.(*UpdateUserEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _User_UpdateUserAvatar_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserAvatarRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServer).UpdateUserAvatar(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.user.User/UpdateUserAvatar",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServer).UpdateUserAvatar(ctx, req.(*UpdateUserAvatarRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// User_ServiceDesc is the grpc.ServiceDesc for User service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var User_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.admin.user.User",
	HandlerType: (*UserServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateUser",
			Handler:    _User_CreateUser_Handler,
		},
		{
			MethodName: "UpdateUser",
			Handler:    _User_UpdateUser_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _User_DeleteUser_Handler,
		},
		{
			MethodName: "GetUser",
			Handler:    _User_GetUser_Handler,
		},
		{
			MethodName: "ListUser",
			Handler:    _User_ListUser_Handler,
		},
		{
			MethodName: "BatchUpdateUserStatus",
			Handler:    _User_BatchUpdateUserStatus_Handler,
		},
		{
			MethodName: "ResetUserPassword",
			Handler:    _User_ResetUserPassword_Handler,
		},
		{
			MethodName: "ResetUserPasswordBySelf",
			Handler:    _User_ResetUserPasswordBySelf_Handler,
		},
		{
			MethodName: "GetUserSelectList",
			Handler:    _User_GetUserSelectList_Handler,
		},
		{
			MethodName: "UpdateUserPhone",
			Handler:    _User_UpdateUserPhone_Handler,
		},
		{
			MethodName: "UpdateUserEmail",
			Handler:    _User_UpdateUserEmail_Handler,
		},
		{
			MethodName: "UpdateUserAvatar",
			Handler:    _User_UpdateUserAvatar_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "admin/user/user.proto",
}
