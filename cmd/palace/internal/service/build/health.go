package build

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/util/ips"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToOperateLogParams(ctx context.Context, menuDo do.Menu, req *middleware.OperateLogParams) *bo.OperateLogParams {
	item := &bo.OperateLogParams{
		Operation:     req.Operation,
		Request:       "",
		Error:         "",
		OriginRequest: "",
		Duration:      req.Duration,
		RequestTime:   req.RequestTime,
		ReplyTime:     req.ReplyTime,
		ClientIP:      ips.GetClientIP(req.OriginRequest),
		UserID:        0,
		TeamID:        0,
		UserAgent:     "",
		UserBaseInfo:  "",
		MenuName:      "",
		MenuID:        0,
	}
	if request, err := json.Marshal(req.Request); validate.IsNil(err) {
		item.Request = string(request)
	} else {
		item.Request = fmt.Sprintf(`{"result": "%v", "error": "%v"}`, req.Request, err)
	}

	if err := req.Error; validate.IsNotNil(err) {
		item.Error = err.Error()
	}

	if validate.IsNotNil(menuDo) {
		item.MenuName = menuDo.GetName()
		item.MenuID = menuDo.GetID()
		if menuDo.GetMenuType().IsMenuTeam() {
			if teamID, ok := permission.GetTeamIDByContext(ctx); ok {
				item.TeamID = teamID
			}
		}
	}

	if userDo, ok := do.GetUserDoContext(ctx); ok && validate.IsNotNil(userDo) {
		item.UserID = userDo.GetID()
		userBase := ToUserBaseItem(userDo)
		userBaseJSON, _ := json.Marshal(userBase)
		item.UserBaseInfo = string(userBaseJSON)
	}

	if request := req.OriginRequest; validate.IsNotNil(request) {
		item.UserAgent = request.UserAgent()
		if originRequest, err := bo.NewHTTPRequest(request); validate.IsNil(err) {
			item.OriginRequest = originRequest.String()
		} else {
			item.OriginRequest = fmt.Sprintf(`{"result": "%v", "error": "%v"}`, request, err)
		}
	}
	return item
}
