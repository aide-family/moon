package bo

import (
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type (
	CreateSysDictBo struct {
		Name     string        `json:"name"`
		Category vobj.Category `json:"category"`
		Status   vobj.Status   `json:"status"`
		Remark   string        `json:"remark"`
		Color    string        `json:"color"`
	}

	UpdateSysDictBo struct {
		ID       uint32        `json:"id"`
		Name     string        `json:"name"`
		Category vobj.Category `json:"category"`
		Status   vobj.Status   `json:"status"`
		Remark   string        `json:"remark"`
		Color    string        `json:"color"`
	}

	ListSysDictBo struct {
		Page Pagination

		Keyword   string        `json:"keyword"`
		Category  vobj.Category `json:"category"`
		Status    vobj.Status   `json:"status"`
		IsDeleted bool          `json:"isDeleted"`
	}

	SelectSysDictBo = ListSysDictBo
)
