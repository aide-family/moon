// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.5.3
// - protoc             v3.19.4
// source: prom/v1/alarm_page.proto

package v1

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

const OperationAlarmPageCreateAlarmPage = "/api.prom.v1.AlarmPage/CreateAlarmPage"
const OperationAlarmPageDeleteAlarmPage = "/api.prom.v1.AlarmPage/DeleteAlarmPage"
const OperationAlarmPageGetAlarmPage = "/api.prom.v1.AlarmPage/GetAlarmPage"
const OperationAlarmPageListAlarmPage = "/api.prom.v1.AlarmPage/ListAlarmPage"
const OperationAlarmPageUpdateAlarmPage = "/api.prom.v1.AlarmPage/UpdateAlarmPage"
const OperationAlarmPageUpdateAlarmPagesStatus = "/api.prom.v1.AlarmPage/UpdateAlarmPagesStatus"

type AlarmPageHTTPServer interface {
	CreateAlarmPage(context.Context, *CreateAlarmPageRequest) (*CreateAlarmPageReply, error)
	DeleteAlarmPage(context.Context, *DeleteAlarmPageRequest) (*DeleteAlarmPageReply, error)
	GetAlarmPage(context.Context, *GetAlarmPageRequest) (*GetAlarmPageReply, error)
	ListAlarmPage(context.Context, *ListAlarmPageRequest) (*ListAlarmPageReply, error)
	UpdateAlarmPage(context.Context, *UpdateAlarmPageRequest) (*UpdateAlarmPageReply, error)
	UpdateAlarmPagesStatus(context.Context, *UpdateAlarmPagesStatusRequest) (*UpdateAlarmPagesStatusReply, error)
}

func RegisterAlarmPageHTTPServer(s *http.Server, srv AlarmPageHTTPServer) {
	r := s.Route("/")
	r.POST("/prom/v1/alarm-page", _AlarmPage_CreateAlarmPage0_HTTP_Handler(srv))
	r.PUT("/prom/v1/alarm-page/{id}", _AlarmPage_UpdateAlarmPage0_HTTP_Handler(srv))
	r.PUT("/prom/v1/alarm-pages/status", _AlarmPage_UpdateAlarmPagesStatus0_HTTP_Handler(srv))
	r.DELETE("/prom/v1/alarm-page/{id}", _AlarmPage_DeleteAlarmPage0_HTTP_Handler(srv))
	r.POST("/prom/v1/alarm-page/{id}", _AlarmPage_GetAlarmPage0_HTTP_Handler(srv))
	r.POST("/prom/v1/alarm-pages", _AlarmPage_ListAlarmPage0_HTTP_Handler(srv))
}

func _AlarmPage_CreateAlarmPage0_HTTP_Handler(srv AlarmPageHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CreateAlarmPageRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAlarmPageCreateAlarmPage)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.CreateAlarmPage(ctx, req.(*CreateAlarmPageRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*CreateAlarmPageReply)
		return ctx.Result(200, reply)
	}
}

func _AlarmPage_UpdateAlarmPage0_HTTP_Handler(srv AlarmPageHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateAlarmPageRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAlarmPageUpdateAlarmPage)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UpdateAlarmPage(ctx, req.(*UpdateAlarmPageRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UpdateAlarmPageReply)
		return ctx.Result(200, reply)
	}
}

func _AlarmPage_UpdateAlarmPagesStatus0_HTTP_Handler(srv AlarmPageHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpdateAlarmPagesStatusRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAlarmPageUpdateAlarmPagesStatus)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UpdateAlarmPagesStatus(ctx, req.(*UpdateAlarmPagesStatusRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UpdateAlarmPagesStatusReply)
		return ctx.Result(200, reply)
	}
}

func _AlarmPage_DeleteAlarmPage0_HTTP_Handler(srv AlarmPageHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in DeleteAlarmPageRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAlarmPageDeleteAlarmPage)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.DeleteAlarmPage(ctx, req.(*DeleteAlarmPageRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*DeleteAlarmPageReply)
		return ctx.Result(200, reply)
	}
}

func _AlarmPage_GetAlarmPage0_HTTP_Handler(srv AlarmPageHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetAlarmPageRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAlarmPageGetAlarmPage)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetAlarmPage(ctx, req.(*GetAlarmPageRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetAlarmPageReply)
		return ctx.Result(200, reply)
	}
}

func _AlarmPage_ListAlarmPage0_HTTP_Handler(srv AlarmPageHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListAlarmPageRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAlarmPageListAlarmPage)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListAlarmPage(ctx, req.(*ListAlarmPageRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ListAlarmPageReply)
		return ctx.Result(200, reply)
	}
}

type AlarmPageHTTPClient interface {
	CreateAlarmPage(ctx context.Context, req *CreateAlarmPageRequest, opts ...http.CallOption) (rsp *CreateAlarmPageReply, err error)
	DeleteAlarmPage(ctx context.Context, req *DeleteAlarmPageRequest, opts ...http.CallOption) (rsp *DeleteAlarmPageReply, err error)
	GetAlarmPage(ctx context.Context, req *GetAlarmPageRequest, opts ...http.CallOption) (rsp *GetAlarmPageReply, err error)
	ListAlarmPage(ctx context.Context, req *ListAlarmPageRequest, opts ...http.CallOption) (rsp *ListAlarmPageReply, err error)
	UpdateAlarmPage(ctx context.Context, req *UpdateAlarmPageRequest, opts ...http.CallOption) (rsp *UpdateAlarmPageReply, err error)
	UpdateAlarmPagesStatus(ctx context.Context, req *UpdateAlarmPagesStatusRequest, opts ...http.CallOption) (rsp *UpdateAlarmPagesStatusReply, err error)
}

type AlarmPageHTTPClientImpl struct {
	cc *http.Client
}

func NewAlarmPageHTTPClient(client *http.Client) AlarmPageHTTPClient {
	return &AlarmPageHTTPClientImpl{client}
}

func (c *AlarmPageHTTPClientImpl) CreateAlarmPage(ctx context.Context, in *CreateAlarmPageRequest, opts ...http.CallOption) (*CreateAlarmPageReply, error) {
	var out CreateAlarmPageReply
	pattern := "/prom/v1/alarm-page"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationAlarmPageCreateAlarmPage))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *AlarmPageHTTPClientImpl) DeleteAlarmPage(ctx context.Context, in *DeleteAlarmPageRequest, opts ...http.CallOption) (*DeleteAlarmPageReply, error) {
	var out DeleteAlarmPageReply
	pattern := "/prom/v1/alarm-page/{id}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationAlarmPageDeleteAlarmPage))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "DELETE", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *AlarmPageHTTPClientImpl) GetAlarmPage(ctx context.Context, in *GetAlarmPageRequest, opts ...http.CallOption) (*GetAlarmPageReply, error) {
	var out GetAlarmPageReply
	pattern := "/prom/v1/alarm-page/{id}"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationAlarmPageGetAlarmPage))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *AlarmPageHTTPClientImpl) ListAlarmPage(ctx context.Context, in *ListAlarmPageRequest, opts ...http.CallOption) (*ListAlarmPageReply, error) {
	var out ListAlarmPageReply
	pattern := "/prom/v1/alarm-pages"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationAlarmPageListAlarmPage))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *AlarmPageHTTPClientImpl) UpdateAlarmPage(ctx context.Context, in *UpdateAlarmPageRequest, opts ...http.CallOption) (*UpdateAlarmPageReply, error) {
	var out UpdateAlarmPageReply
	pattern := "/prom/v1/alarm-page/{id}"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationAlarmPageUpdateAlarmPage))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "PUT", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *AlarmPageHTTPClientImpl) UpdateAlarmPagesStatus(ctx context.Context, in *UpdateAlarmPagesStatusRequest, opts ...http.CallOption) (*UpdateAlarmPagesStatusReply, error) {
	var out UpdateAlarmPagesStatusReply
	pattern := "/prom/v1/alarm-pages/status"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationAlarmPageUpdateAlarmPagesStatus))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "PUT", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
