// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.5.3
// - protoc             v3.19.4
// source: alarm/history/history.proto

package history

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

const OperationHistoryGetHistory = "/api.alarm.history.History/GetHistory"
const OperationHistoryListHistory = "/api.alarm.history.History/ListHistory"

type HistoryHTTPServer interface {
	GetHistory(context.Context, *GetHistoryRequest) (*GetHistoryReply, error)
	ListHistory(context.Context, *ListHistoryRequest) (*ListHistoryReply, error)
}

func RegisterHistoryHTTPServer(s *http.Server, srv HistoryHTTPServer) {
	r := s.Route("/")
	r.POST("/api/v1/alarm/history/get", _History_GetHistory0_HTTP_Handler(srv))
	r.POST("/api/v1/alarm/history/list", _History_ListHistory0_HTTP_Handler(srv))
}

func _History_GetHistory0_HTTP_Handler(srv HistoryHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetHistoryRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationHistoryGetHistory)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetHistory(ctx, req.(*GetHistoryRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetHistoryReply)
		return ctx.Result(200, reply)
	}
}

func _History_ListHistory0_HTTP_Handler(srv HistoryHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListHistoryRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationHistoryListHistory)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListHistory(ctx, req.(*ListHistoryRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ListHistoryReply)
		return ctx.Result(200, reply)
	}
}

type HistoryHTTPClient interface {
	GetHistory(ctx context.Context, req *GetHistoryRequest, opts ...http.CallOption) (rsp *GetHistoryReply, err error)
	ListHistory(ctx context.Context, req *ListHistoryRequest, opts ...http.CallOption) (rsp *ListHistoryReply, err error)
}

type HistoryHTTPClientImpl struct {
	cc *http.Client
}

func NewHistoryHTTPClient(client *http.Client) HistoryHTTPClient {
	return &HistoryHTTPClientImpl{client}
}

func (c *HistoryHTTPClientImpl) GetHistory(ctx context.Context, in *GetHistoryRequest, opts ...http.CallOption) (*GetHistoryReply, error) {
	var out GetHistoryReply
	pattern := "/api/v1/alarm/history/get"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationHistoryGetHistory))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *HistoryHTTPClientImpl) ListHistory(ctx context.Context, in *ListHistoryRequest, opts ...http.CallOption) (*ListHistoryReply, error) {
	var out ListHistoryReply
	pattern := "/api/v1/alarm/history/list"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationHistoryListHistory))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
