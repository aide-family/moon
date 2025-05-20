package build

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToOperateLogParams(ctx context.Context, req *middleware.OperateLogParams) *bo.OperateLogParams {
	item := &bo.OperateLogParams{
		Operation:     req.Operation,
		Request:       "",
		Reply:         "",
		Error:         "",
		OriginRequest: "",
		Duration:      req.Duration,
		RequestTime:   req.RequestTime,
		ReplyTime:     req.ReplyTime,
		ClientIP:      "",
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
		item.Request = fmt.Sprintf(`{"result": "%v"}`, req.Request)
	}
	if reply, err := json.Marshal(req.Reply); validate.IsNil(err) {
		item.Reply = string(reply)
	} else {
		item.Reply = fmt.Sprintf(`{"result": "%v"}`, req.Reply)
	}

	if err := req.Error; validate.IsNotNil(err) {
		item.Error = err.Error()
	}

	if menuDo, ok := do.GetMenuDoContext(ctx); ok && validate.IsNotNil(menuDo) {
		item.MenuName = menuDo.GetName()
		item.MenuID = menuDo.GetID()
	}

	if userDo, ok := do.GetUserDoContext(ctx); ok && validate.IsNotNil(userDo) {
		item.UserID = userDo.GetID()
		userBase := ToUserBaseItem(userDo)
		userBaseJson, _ := json.Marshal(userBase)
		item.UserBaseInfo = string(userBaseJson)
	}

	if teamID, ok := permission.GetTeamIDByContext(ctx); ok {
		item.TeamID = teamID
	}

	if request := req.OriginRequest; validate.IsNotNil(request) {
		item.ClientIP = request.RemoteAddr
		item.UserAgent = request.UserAgent()
		if originRequest, err := json.Marshal(request); validate.IsNil(err) {
			item.OriginRequest = string(originRequest)
		} else {
			item.OriginRequest = fmt.Sprintf(`{"result": "%v"}`, request)
		}
	}
	return item
}
