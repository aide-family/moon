package bo

import (
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type (
	ListDashboardReq struct {
		Page    Pagination  `json:"page"`
		Keyword string      `json:"keyword"`
		Status  vobj.Status `json:"status"`
	}

	CreateMyDashboardBO struct {
		Status vobj.Status  `json:"status"`
		Remark string       `json:"remark"`
		Title  string       `json:"title"`
		Color  string       `json:"color"`
		UserId uint32       `json:"userId"`
		Charts []*MyChartBO `json:"charts"`
	}

	UpdateMyDashboardBO struct {
		Id uint32 `json:"id"`
		CreateMyDashboardBO
	}
)
