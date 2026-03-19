package memberv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	memberdomain "github.com/aide-family/goddess/domain/member"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	memberdomain.RegisterMemberV1Factory(config.DomainConfig_OUTER, NewOuterMember)
}

// NewOuterMember creates a member client that calls a remote goddess (OUTER driver).
func NewOuterMember(c *config.DomainConfig, driver *anypb.Any) (goddessv1.MemberServer, func() error, error) {
	outer := &config.OuterServerConfig{}
	if pointer.IsNotNil(driver) {
		if err := anypb.UnmarshalTo(driver, outer, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal outer server config failed: %v", err)
		}
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("goddess.member", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewMemberHTTPClient(httpClient)
		return &outerMemberServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := goddessv1.NewMemberClient(grpcConn)
		return &outerMemberServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerMemberServer struct {
	goddessv1.UnimplementedMemberServer

	httpClient goddessv1.MemberHTTPClient
	grpcClient goddessv1.MemberClient
}

func (o *outerMemberServer) ListMember(ctx context.Context, req *goddessv1.ListMemberRequest) (*goddessv1.ListMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListMember(ctx, req)
	}
	return o.grpcClient.ListMember(ctx, req)
}

func (o *outerMemberServer) GetMember(ctx context.Context, req *goddessv1.GetMemberRequest) (*goddessv1.MemberItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetMember(ctx, req)
	}
	return o.grpcClient.GetMember(ctx, req)
}

func (o *outerMemberServer) SelectMember(ctx context.Context, req *goddessv1.SelectMemberRequest) (*goddessv1.SelectMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectMember(ctx, req)
	}
	return o.grpcClient.SelectMember(ctx, req)
}

func (o *outerMemberServer) InviteMember(ctx context.Context, req *goddessv1.InviteMemberRequest) (*goddessv1.InviteMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.InviteMember(ctx, req)
	}
	return o.grpcClient.InviteMember(ctx, req)
}

func (o *outerMemberServer) DismissMember(ctx context.Context, req *goddessv1.DismissMemberRequest) (*goddessv1.DismissMemberReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DismissMember(ctx, req)
	}
	return o.grpcClient.DismissMember(ctx, req)
}

func (o *outerMemberServer) UpdateMemberStatus(ctx context.Context, req *goddessv1.UpdateMemberStatusRequest) (*goddessv1.UpdateMemberStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateMemberStatus(ctx, req)
	}
	return o.grpcClient.UpdateMemberStatus(ctx, req)
}
