// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.7.3
// - protoc             v3.19.4
// source: server/prom/notify/chat_group.proto

package notify

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationChatGroupCreateChatGroup = "/api.server.prom.notify.ChatGroup/CreateChatGroup"
const OperationChatGroupDeleteChatGroup = "/api.server.prom.notify.ChatGroup/DeleteChatGroup"
const OperationChatGroupGetChatGroup = "/api.server.prom.notify.ChatGroup/GetChatGroup"
const OperationChatGroupListChatGroup = "/api.server.prom.notify.ChatGroup/ListChatGroup"
const OperationChatGroupSelectChatGroup = "/api.server.prom.notify.ChatGroup/SelectChatGroup"
const OperationChatGroupUpdateChatGroup = "/api.server.prom.notify.ChatGroup/UpdateChatGroup"

type ChatGroupHTTPServer interface {
	// CreateChatGroup 创建通知群组
	CreateChatGroup(context.Context, *CreateChatGroupRequest) (*CreateChatGroupReply, error)
	// DeleteChatGroup 删除通知群组
	DeleteChatGroup(context.Context, *DeleteChatGroupRequest) (*DeleteChatGroupReply, error)
	// GetChatGroup 获取通知群组
	GetChatGroup(context.Context, *GetChatGroupRequest) (*GetChatGroupReply, error)
	// ListChatGroup 获取通知群组列表
	ListChatGroup(context.Context, *ListChatGroupRequest) (*ListChatGroupReply, error)
	// SelectChatGroup 获取通知群组列表(下拉选择)
	SelectChatGroup(context.Context, *SelectChatGroupRequest) (*SelectChatGroupReply, error)
	// UpdateChatGroup 更新通知群组
	UpdateChatGroup(context.Context, *UpdateChatGroupRequest) (*UpdateChatGroupReply, error)
}

func RegisterChatGroupHTTPServer(s *http.Server, srv ChatGroupHTTPServer) {
	r := s.Route("/")
	r.POST("/api/v1/chat/group/create", _ChatGroup_CreateChatGroup0_HTTP_Handler(srv))
	r.POST("/api/v1/chat/group/update", _ChatGroup_UpdateChatGroup0_HTTP_Handler(srv))
	r.POST("/api/v1/chat/group/delete", _ChatGroup_DeleteChatGroup0_HTTP_Handler(srv))
	r.POST("/api/v1/chat/group/get", _ChatGroup_GetChatGroup0_HTTP_Handler(srv))
	r.POST("/api/v1/chat/group/list", _ChatGroup_ListChatGroup0_HTTP_Handler(srv))
	r.POST("/api/v1/chat/group/select", _ChatGroup_SelectChatGroup0_HTTP_Handler(srv))
}

func _ChatGroup_CreateChatGroup0_HTTP_Handler(srv ChatGroupHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CreateChatGroupRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationChatGroupCreateChatGroup)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.CreateChatGroup(ctx, req.(*CreateChatGroupRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*CreateChatGroupReply)
		return ctx.Result(200, reply)
	}
}

func _ChatGroup_UpdateChatGroup0_HTTP_Handler(srv ChatGroupHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateChatGroupRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationChatGroupUpdateChatGroup)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UpdateChatGroup(ctx, req.(*UpdateChatGroupRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UpdateChatGroupReply)
		return ctx.Result(200, reply)
	}
}

func _ChatGroup_DeleteChatGroup0_HTTP_Handler(srv ChatGroupHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in DeleteChatGroupRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationChatGroupDeleteChatGroup)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.DeleteChatGroup(ctx, req.(*DeleteChatGroupRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*DeleteChatGroupReply)
		return ctx.Result(200, reply)
	}
}

func _ChatGroup_GetChatGroup0_HTTP_Handler(srv ChatGroupHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetChatGroupRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationChatGroupGetChatGroup)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetChatGroup(ctx, req.(*GetChatGroupRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetChatGroupReply)
		return ctx.Result(200, reply)
	}
}

func _ChatGroup_ListChatGroup0_HTTP_Handler(srv ChatGroupHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListChatGroupRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationChatGroupListChatGroup)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListChatGroup(ctx, req.(*ListChatGroupRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ListChatGroupReply)
		return ctx.Result(200, reply)
	}
}

func _ChatGroup_SelectChatGroup0_HTTP_Handler(srv ChatGroupHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in SelectChatGroupRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationChatGroupSelectChatGroup)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.SelectChatGroup(ctx, req.(*SelectChatGroupRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*SelectChatGroupReply)
		return ctx.Result(200, reply)
	}
}

type ChatGroupHTTPClient interface {
	CreateChatGroup(ctx context.Context, req *CreateChatGroupRequest, opts ...http.CallOption) (rsp *CreateChatGroupReply, err error)
	DeleteChatGroup(ctx context.Context, req *DeleteChatGroupRequest, opts ...http.CallOption) (rsp *DeleteChatGroupReply, err error)
	GetChatGroup(ctx context.Context, req *GetChatGroupRequest, opts ...http.CallOption) (rsp *GetChatGroupReply, err error)
	ListChatGroup(ctx context.Context, req *ListChatGroupRequest, opts ...http.CallOption) (rsp *ListChatGroupReply, err error)
	SelectChatGroup(ctx context.Context, req *SelectChatGroupRequest, opts ...http.CallOption) (rsp *SelectChatGroupReply, err error)
	UpdateChatGroup(ctx context.Context, req *UpdateChatGroupRequest, opts ...http.CallOption) (rsp *UpdateChatGroupReply, err error)
}

type ChatGroupHTTPClientImpl struct {
	cc *http.Client
}

func NewChatGroupHTTPClient(client *http.Client) ChatGroupHTTPClient {
	return &ChatGroupHTTPClientImpl{client}
}

func (c *ChatGroupHTTPClientImpl) CreateChatGroup(ctx context.Context, in *CreateChatGroupRequest, opts ...http.CallOption) (*CreateChatGroupReply, error) {
	var out CreateChatGroupReply
	pattern := "/api/v1/chat/group/create"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationChatGroupCreateChatGroup))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ChatGroupHTTPClientImpl) DeleteChatGroup(ctx context.Context, in *DeleteChatGroupRequest, opts ...http.CallOption) (*DeleteChatGroupReply, error) {
	var out DeleteChatGroupReply
	pattern := "/api/v1/chat/group/delete"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationChatGroupDeleteChatGroup))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ChatGroupHTTPClientImpl) GetChatGroup(ctx context.Context, in *GetChatGroupRequest, opts ...http.CallOption) (*GetChatGroupReply, error) {
	var out GetChatGroupReply
	pattern := "/api/v1/chat/group/get"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationChatGroupGetChatGroup))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ChatGroupHTTPClientImpl) ListChatGroup(ctx context.Context, in *ListChatGroupRequest, opts ...http.CallOption) (*ListChatGroupReply, error) {
	var out ListChatGroupReply
	pattern := "/api/v1/chat/group/list"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationChatGroupListChatGroup))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ChatGroupHTTPClientImpl) SelectChatGroup(ctx context.Context, in *SelectChatGroupRequest, opts ...http.CallOption) (*SelectChatGroupReply, error) {
	var out SelectChatGroupReply
	pattern := "/api/v1/chat/group/select"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationChatGroupSelectChatGroup))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ChatGroupHTTPClientImpl) UpdateChatGroup(ctx context.Context, in *UpdateChatGroupRequest, opts ...http.CallOption) (*UpdateChatGroupReply, error) {
	var out UpdateChatGroupReply
	pattern := "/api/v1/chat/group/update"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationChatGroupUpdateChatGroup))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
