// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: admin/resource/resource.proto

package resource

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

// ResourceClient is the client API for Resource service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ResourceClient interface {
	// 获取资源详情
	GetResource(ctx context.Context, in *GetResourceRequest, opts ...grpc.CallOption) (*GetResourceReply, error)
	// 获取资源列表
	ListResource(ctx context.Context, in *ListResourceRequest, opts ...grpc.CallOption) (*ListResourceReply, error)
	// 批量更新资源状态
	BatchUpdateResourceStatus(ctx context.Context, in *BatchUpdateResourceStatusRequest, opts ...grpc.CallOption) (*BatchUpdateResourceStatusReply, error)
	// 获取资源下拉列表
	GetResourceSelectList(ctx context.Context, in *GetResourceSelectListRequest, opts ...grpc.CallOption) (*GetResourceSelectListReply, error)
}

type resourceClient struct {
	cc grpc.ClientConnInterface
}

func NewResourceClient(cc grpc.ClientConnInterface) ResourceClient {
	return &resourceClient{cc}
}

func (c *resourceClient) GetResource(ctx context.Context, in *GetResourceRequest, opts ...grpc.CallOption) (*GetResourceReply, error) {
	out := new(GetResourceReply)
	err := c.cc.Invoke(ctx, "/api.admin.resource.Resource/GetResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceClient) ListResource(ctx context.Context, in *ListResourceRequest, opts ...grpc.CallOption) (*ListResourceReply, error) {
	out := new(ListResourceReply)
	err := c.cc.Invoke(ctx, "/api.admin.resource.Resource/ListResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceClient) BatchUpdateResourceStatus(ctx context.Context, in *BatchUpdateResourceStatusRequest, opts ...grpc.CallOption) (*BatchUpdateResourceStatusReply, error) {
	out := new(BatchUpdateResourceStatusReply)
	err := c.cc.Invoke(ctx, "/api.admin.resource.Resource/BatchUpdateResourceStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *resourceClient) GetResourceSelectList(ctx context.Context, in *GetResourceSelectListRequest, opts ...grpc.CallOption) (*GetResourceSelectListReply, error) {
	out := new(GetResourceSelectListReply)
	err := c.cc.Invoke(ctx, "/api.admin.resource.Resource/GetResourceSelectList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ResourceServer is the server API for Resource service.
// All implementations must embed UnimplementedResourceServer
// for forward compatibility
type ResourceServer interface {
	// 获取资源详情
	GetResource(context.Context, *GetResourceRequest) (*GetResourceReply, error)
	// 获取资源列表
	ListResource(context.Context, *ListResourceRequest) (*ListResourceReply, error)
	// 批量更新资源状态
	BatchUpdateResourceStatus(context.Context, *BatchUpdateResourceStatusRequest) (*BatchUpdateResourceStatusReply, error)
	// 获取资源下拉列表
	GetResourceSelectList(context.Context, *GetResourceSelectListRequest) (*GetResourceSelectListReply, error)
	mustEmbedUnimplementedResourceServer()
}

// UnimplementedResourceServer must be embedded to have forward compatible implementations.
type UnimplementedResourceServer struct {
}

func (UnimplementedResourceServer) GetResource(context.Context, *GetResourceRequest) (*GetResourceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResource not implemented")
}
func (UnimplementedResourceServer) ListResource(context.Context, *ListResourceRequest) (*ListResourceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListResource not implemented")
}
func (UnimplementedResourceServer) BatchUpdateResourceStatus(context.Context, *BatchUpdateResourceStatusRequest) (*BatchUpdateResourceStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchUpdateResourceStatus not implemented")
}
func (UnimplementedResourceServer) GetResourceSelectList(context.Context, *GetResourceSelectListRequest) (*GetResourceSelectListReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResourceSelectList not implemented")
}
func (UnimplementedResourceServer) mustEmbedUnimplementedResourceServer() {}

// UnsafeResourceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ResourceServer will
// result in compilation errors.
type UnsafeResourceServer interface {
	mustEmbedUnimplementedResourceServer()
}

func RegisterResourceServer(s grpc.ServiceRegistrar, srv ResourceServer) {
	s.RegisterService(&Resource_ServiceDesc, srv)
}

func _Resource_GetResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceServer).GetResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.resource.Resource/GetResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceServer).GetResource(ctx, req.(*GetResourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Resource_ListResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListResourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceServer).ListResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.resource.Resource/ListResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceServer).ListResource(ctx, req.(*ListResourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Resource_BatchUpdateResourceStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchUpdateResourceStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceServer).BatchUpdateResourceStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.resource.Resource/BatchUpdateResourceStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceServer).BatchUpdateResourceStatus(ctx, req.(*BatchUpdateResourceStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Resource_GetResourceSelectList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResourceSelectListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ResourceServer).GetResourceSelectList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.admin.resource.Resource/GetResourceSelectList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ResourceServer).GetResourceSelectList(ctx, req.(*GetResourceSelectListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Resource_ServiceDesc is the grpc.ServiceDesc for Resource service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Resource_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.admin.resource.Resource",
	HandlerType: (*ResourceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetResource",
			Handler:    _Resource_GetResource_Handler,
		},
		{
			MethodName: "ListResource",
			Handler:    _Resource_ListResource_Handler,
		},
		{
			MethodName: "BatchUpdateResourceStatus",
			Handler:    _Resource_BatchUpdateResourceStatus_Handler,
		},
		{
			MethodName: "GetResourceSelectList",
			Handler:    _Resource_GetResourceSelectList_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "admin/resource/resource.proto",
}
