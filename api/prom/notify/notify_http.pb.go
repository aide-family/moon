// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.5.3
// - protoc             v3.19.4
// source: prom/notify/notify.proto

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

const OperationNotifyCreateNotify = "/api.prom.notify.Notify/CreateNotify"
const OperationNotifyDeleteNotify = "/api.prom.notify.Notify/DeleteNotify"
const OperationNotifyGetNotify = "/api.prom.notify.Notify/GetNotify"
const OperationNotifyListNotify = "/api.prom.notify.Notify/ListNotify"
const OperationNotifySelectNotify = "/api.prom.notify.Notify/SelectNotify"
const OperationNotifyUpdateNotify = "/api.prom.notify.Notify/UpdateNotify"

type NotifyHTTPServer interface {
	CreateNotify(context.Context, *CreateNotifyRequest) (*CreateNotifyReply, error)
	DeleteNotify(context.Context, *DeleteNotifyRequest) (*DeleteNotifyReply, error)
	GetNotify(context.Context, *GetNotifyRequest) (*GetNotifyReply, error)
	ListNotify(context.Context, *ListNotifyRequest) (*ListNotifyReply, error)
	SelectNotify(context.Context, *SelectNotifyRequest) (*SelectNotifyReply, error)
	UpdateNotify(context.Context, *UpdateNotifyRequest) (*UpdateNotifyReply, error)
}

func RegisterNotifyHTTPServer(s *http.Server, srv NotifyHTTPServer) {
	r := s.Route("/")
	r.POST("/api/v1/prom/notify/create", _Notify_CreateNotify0_HTTP_Handler(srv))
	r.POST("/api/v1/prom/notify/update", _Notify_UpdateNotify0_HTTP_Handler(srv))
	r.POST("/api/v1/prom/notify/delete", _Notify_DeleteNotify0_HTTP_Handler(srv))
	r.GET("/api/v1/prom/notify/get", _Notify_GetNotify0_HTTP_Handler(srv))
	r.POST("/api/v1/prom/notify/list", _Notify_ListNotify0_HTTP_Handler(srv))
	r.POST("/api/v1/prom/notify/select", _Notify_SelectNotify0_HTTP_Handler(srv))
}

func _Notify_CreateNotify0_HTTP_Handler(srv NotifyHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CreateNotifyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNotifyCreateNotify)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.CreateNotify(ctx, req.(*CreateNotifyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*CreateNotifyReply)
		return ctx.Result(200, reply)
	}
}

func _Notify_UpdateNotify0_HTTP_Handler(srv NotifyHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateNotifyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNotifyUpdateNotify)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UpdateNotify(ctx, req.(*UpdateNotifyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UpdateNotifyReply)
		return ctx.Result(200, reply)
	}
}

func _Notify_DeleteNotify0_HTTP_Handler(srv NotifyHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in DeleteNotifyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNotifyDeleteNotify)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.DeleteNotify(ctx, req.(*DeleteNotifyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*DeleteNotifyReply)
		return ctx.Result(200, reply)
	}
}

func _Notify_GetNotify0_HTTP_Handler(srv NotifyHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetNotifyRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNotifyGetNotify)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetNotify(ctx, req.(*GetNotifyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetNotifyReply)
		return ctx.Result(200, reply)
	}
}

func _Notify_ListNotify0_HTTP_Handler(srv NotifyHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListNotifyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNotifyListNotify)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListNotify(ctx, req.(*ListNotifyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ListNotifyReply)
		return ctx.Result(200, reply)
	}
}

func _Notify_SelectNotify0_HTTP_Handler(srv NotifyHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in SelectNotifyRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNotifySelectNotify)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.SelectNotify(ctx, req.(*SelectNotifyRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*SelectNotifyReply)
		return ctx.Result(200, reply)
	}
}

type NotifyHTTPClient interface {
	CreateNotify(ctx context.Context, req *CreateNotifyRequest, opts ...http.CallOption) (rsp *CreateNotifyReply, err error)
	DeleteNotify(ctx context.Context, req *DeleteNotifyRequest, opts ...http.CallOption) (rsp *DeleteNotifyReply, err error)
	GetNotify(ctx context.Context, req *GetNotifyRequest, opts ...http.CallOption) (rsp *GetNotifyReply, err error)
	ListNotify(ctx context.Context, req *ListNotifyRequest, opts ...http.CallOption) (rsp *ListNotifyReply, err error)
	SelectNotify(ctx context.Context, req *SelectNotifyRequest, opts ...http.CallOption) (rsp *SelectNotifyReply, err error)
	UpdateNotify(ctx context.Context, req *UpdateNotifyRequest, opts ...http.CallOption) (rsp *UpdateNotifyReply, err error)
}

type NotifyHTTPClientImpl struct {
	cc *http.Client
}

func NewNotifyHTTPClient(client *http.Client) NotifyHTTPClient {
	return &NotifyHTTPClientImpl{client}
}

func (c *NotifyHTTPClientImpl) CreateNotify(ctx context.Context, in *CreateNotifyRequest, opts ...http.CallOption) (*CreateNotifyReply, error) {
	var out CreateNotifyReply
	pattern := "/api/v1/prom/notify/create"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNotifyCreateNotify))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *NotifyHTTPClientImpl) DeleteNotify(ctx context.Context, in *DeleteNotifyRequest, opts ...http.CallOption) (*DeleteNotifyReply, error) {
	var out DeleteNotifyReply
	pattern := "/api/v1/prom/notify/delete"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNotifyDeleteNotify))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *NotifyHTTPClientImpl) GetNotify(ctx context.Context, in *GetNotifyRequest, opts ...http.CallOption) (*GetNotifyReply, error) {
	var out GetNotifyReply
	pattern := "/api/v1/prom/notify/get"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationNotifyGetNotify))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *NotifyHTTPClientImpl) ListNotify(ctx context.Context, in *ListNotifyRequest, opts ...http.CallOption) (*ListNotifyReply, error) {
	var out ListNotifyReply
	pattern := "/api/v1/prom/notify/list"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNotifyListNotify))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *NotifyHTTPClientImpl) SelectNotify(ctx context.Context, in *SelectNotifyRequest, opts ...http.CallOption) (*SelectNotifyReply, error) {
	var out SelectNotifyReply
	pattern := "/api/v1/prom/notify/select"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNotifySelectNotify))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *NotifyHTTPClientImpl) UpdateNotify(ctx context.Context, in *UpdateNotifyRequest, opts ...http.CallOption) (*UpdateNotifyReply, error) {
	var out UpdateNotifyReply
	pattern := "/api/v1/prom/notify/update"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNotifyUpdateNotify))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
