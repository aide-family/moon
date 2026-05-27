package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewOuterMember creates a member client that calls a remote goddess (external domain).
func newGoddessMember(c *config.ExternalDomainConfig) (goddessv1.MemberServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.member", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewMemberHTTPClient(httpClient)
		return &outerMemberServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewMemberClient(grpcConn)
		return &outerMemberServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerMemberServer struct {
	goddessv1.UnimplementedMemberServer

	cfg        *config.ExternalDomainConfig
	httpClient goddessv1.MemberHTTPClient
	grpcClient goddessv1.MemberClient
}

func (o *outerMemberServer) ListMember(ctx context.Context, req *goddessv1.ListMemberRequest) (*goddessv1.ListMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListMember(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.ListMember(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}

func (o *outerMemberServer) GetMember(ctx context.Context, req *goddessv1.GetMemberRequest) (*goddessv1.MemberItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetMember(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.GetMember(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}

func (o *outerMemberServer) GetMemberByUserUID(ctx context.Context, req *goddessv1.GetMemberByUserUIDRequest) (*goddessv1.MemberItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetMemberByUserUID(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.GetMemberByUserUID(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}

func (o *outerMemberServer) SelectMember(ctx context.Context, req *goddessv1.SelectMemberRequest) (*goddessv1.SelectMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectMember(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.SelectMember(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}

func (o *outerMemberServer) InviteMember(ctx context.Context, req *goddessv1.InviteMemberRequest) (*goddessv1.InviteMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.InviteMember(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.InviteMember(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}

func (o *outerMemberServer) DismissMember(ctx context.Context, req *goddessv1.DismissMemberRequest) (*goddessv1.DismissMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DismissMember(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.DismissMember(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}

func (o *outerMemberServer) UpdateMemberStatus(ctx context.Context, req *goddessv1.UpdateMemberStatusRequest) (*goddessv1.UpdateMemberStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateMemberStatus(externalContext(ctx, o.cfg), req)
	}
	if o.grpcClient != nil {
		return o.grpcClient.UpdateMemberStatus(externalContext(ctx, o.cfg), req)
	}
	return nil, merr.ErrorInternalServer("no client available")
}
