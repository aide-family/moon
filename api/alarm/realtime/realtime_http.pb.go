// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.5.3
// - protoc             v3.19.4
// source: alarm/realtime/realtime.proto

package realtime

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

const OperationRealtimeGetRealtime = "/api.alarm.realtime.Realtime/GetRealtime"
const OperationRealtimeIntervene = "/api.alarm.realtime.Realtime/Intervene"
const OperationRealtimeListRealtime = "/api.alarm.realtime.Realtime/ListRealtime"
const OperationRealtimeSuppress = "/api.alarm.realtime.Realtime/Suppress"
const OperationRealtimeUpgrade = "/api.alarm.realtime.Realtime/Upgrade"

type RealtimeHTTPServer interface {
	GetRealtime(context.Context, *GetRealtimeRequest) (*GetRealtimeReply, error)
	Intervene(context.Context, *InterveneRequest) (*InterveneReply, error)
	ListRealtime(context.Context, *ListRealtimeRequest) (*ListRealtimeReply, error)
	Suppress(context.Context, *SuppressRequest) (*SuppressReply, error)
	Upgrade(context.Context, *UpgradeRequest) (*UpgradeReply, error)
}

func RegisterRealtimeHTTPServer(s *http.Server, srv RealtimeHTTPServer) {
	r := s.Route("/")
	r.GET("/api/v1/alarm/realtime/detail", _Realtime_GetRealtime0_HTTP_Handler(srv))
	r.POST("/api/v1/alarm/realtime/list", _Realtime_ListRealtime0_HTTP_Handler(srv))
	r.POST("/api/v1/alarm/realtime/intervene", _Realtime_Intervene0_HTTP_Handler(srv))
	r.POST("/api/v1/alarm/realtime/upgrade", _Realtime_Upgrade0_HTTP_Handler(srv))
	r.POST("/api/v1/alarm/realtime/suppress", _Realtime_Suppress0_HTTP_Handler(srv))
}

func _Realtime_GetRealtime0_HTTP_Handler(srv RealtimeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetRealtimeRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationRealtimeGetRealtime)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetRealtime(ctx, req.(*GetRealtimeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetRealtimeReply)
		return ctx.Result(200, reply)
	}
}

func _Realtime_ListRealtime0_HTTP_Handler(srv RealtimeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in ListRealtimeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationRealtimeListRealtime)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.ListRealtime(ctx, req.(*ListRealtimeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*ListRealtimeReply)
		return ctx.Result(200, reply)
	}
}

func _Realtime_Intervene0_HTTP_Handler(srv RealtimeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in InterveneRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationRealtimeIntervene)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Intervene(ctx, req.(*InterveneRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*InterveneReply)
		return ctx.Result(200, reply)
	}
}

func _Realtime_Upgrade0_HTTP_Handler(srv RealtimeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UpgradeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationRealtimeUpgrade)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Upgrade(ctx, req.(*UpgradeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UpgradeReply)
		return ctx.Result(200, reply)
	}
}

func _Realtime_Suppress0_HTTP_Handler(srv RealtimeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in SuppressRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationRealtimeSuppress)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Suppress(ctx, req.(*SuppressRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*SuppressReply)
		return ctx.Result(200, reply)
	}
}

type RealtimeHTTPClient interface {
	GetRealtime(ctx context.Context, req *GetRealtimeRequest, opts ...http.CallOption) (rsp *GetRealtimeReply, err error)
	Intervene(ctx context.Context, req *InterveneRequest, opts ...http.CallOption) (rsp *InterveneReply, err error)
	ListRealtime(ctx context.Context, req *ListRealtimeRequest, opts ...http.CallOption) (rsp *ListRealtimeReply, err error)
	Suppress(ctx context.Context, req *SuppressRequest, opts ...http.CallOption) (rsp *SuppressReply, err error)
	Upgrade(ctx context.Context, req *UpgradeRequest, opts ...http.CallOption) (rsp *UpgradeReply, err error)
}

type RealtimeHTTPClientImpl struct {
	cc *http.Client
}

func NewRealtimeHTTPClient(client *http.Client) RealtimeHTTPClient {
	return &RealtimeHTTPClientImpl{client}
}

func (c *RealtimeHTTPClientImpl) GetRealtime(ctx context.Context, in *GetRealtimeRequest, opts ...http.CallOption) (*GetRealtimeReply, error) {
	var out GetRealtimeReply
	pattern := "/api/v1/alarm/realtime/detail"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationRealtimeGetRealtime))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *RealtimeHTTPClientImpl) Intervene(ctx context.Context, in *InterveneRequest, opts ...http.CallOption) (*InterveneReply, error) {
	var out InterveneReply
	pattern := "/api/v1/alarm/realtime/intervene"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationRealtimeIntervene))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *RealtimeHTTPClientImpl) ListRealtime(ctx context.Context, in *ListRealtimeRequest, opts ...http.CallOption) (*ListRealtimeReply, error) {
	var out ListRealtimeReply
	pattern := "/api/v1/alarm/realtime/list"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationRealtimeListRealtime))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *RealtimeHTTPClientImpl) Suppress(ctx context.Context, in *SuppressRequest, opts ...http.CallOption) (*SuppressReply, error) {
	var out SuppressReply
	pattern := "/api/v1/alarm/realtime/suppress"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationRealtimeSuppress))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *RealtimeHTTPClientImpl) Upgrade(ctx context.Context, in *UpgradeRequest, opts ...http.CallOption) (*UpgradeReply, error) {
	var out UpgradeReply
	pattern := "/api/v1/alarm/realtime/upgrade"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationRealtimeUpgrade))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
