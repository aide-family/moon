package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToServerRegisterReq(req *common.ServerRegisterRequest) *bo.ServerRegisterReq {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.ServerRegisterReq{
		ServerType: vobj.ServerType(req.GetServerType()),
		Server:     req.GetServer(),
		Discovery:  req.GetDiscovery(),
		TeamIds:    req.GetTeamIds(),
		IsOnline:   req.GetIsOnline(),
		Uuid:       req.GetUuid(),
	}
}
