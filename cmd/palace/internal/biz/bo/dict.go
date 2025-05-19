package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

type Dict interface {
	GetID() uint32
	GetKey() string
	GetValue() string
	GetStatus() vobj.GlobalStatus
	GetType() vobj.DictType
	GetColor() string
	GetLang() string
}

type SaveDictReq struct {
	dictItem Dict

	DictID uint32            `json:"dictID"`
	Key    string            `json:"key"`
	Value  string            `json:"value"`
	Status vobj.GlobalStatus `json:"status"`
	Type   vobj.DictType     `json:"type"`
	Color  string            `json:"color"`
	Lang   string            `json:"lang"`
}

func (s *SaveDictReq) GetID() uint32 {
	if s == nil {
		return 0
	}
	if s.dictItem == nil {
		return s.DictID
	}
	return s.dictItem.GetID()
}

func (s *SaveDictReq) GetKey() string {
	return s.Key
}

func (s *SaveDictReq) GetValue() string {
	return s.Value
}

func (s *SaveDictReq) GetStatus() vobj.GlobalStatus {
	return s.Status
}

func (s *SaveDictReq) GetType() vobj.DictType {
	return s.Type
}

func (s *SaveDictReq) GetColor() string {
	return s.Color
}

func (s *SaveDictReq) GetLang() string {
	return s.Lang
}

func (s *SaveDictReq) WithUpdateParams(dictItem Dict) Dict {
	s.dictItem = dictItem
	return s
}

type UpdateDictStatusReq struct {
	DictIds []uint32          `json:"dictIds"`
	Status  vobj.GlobalStatus `json:"status"`
}

type OperateOneDictReq struct {
	DictID uint32 `json:"dictID"`
}

type ListDictReq struct {
	*PaginationRequest
	DictTypes []vobj.DictType   `json:"dictTypes"`
	Status    vobj.GlobalStatus `json:"status"`
	Keyword   string            `json:"keyword"`
	Langs     []string          `json:"langs"`
}

func (r *ListDictReq) ToListReply(dictItems []do.TeamDict) *ListDictReply {
	return &ListDictReply{
		PaginationReply: r.ToReply(),
		Items:           dictItems,
	}
}

type ListDictReply = ListReply[do.TeamDict]

type SelectDictReq struct {
	*PaginationRequest
	DictTypes []vobj.DictType   `json:"dictTypes"`
	Status    vobj.GlobalStatus `json:"status"`
	Keyword   string            `json:"keyword"`
	Langs     []string          `json:"langs"`
}

func (r *SelectDictReq) ToSelectReply(dictItems []do.TeamDict) *SelectDictReply {
	return &SelectDictReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(dictItems, func(dict do.TeamDict) SelectItem {
			return &selectItem{
				Label:    dict.GetValue(),
				Disabled: dict.GetDeletedAt() > 0 || !dict.GetStatus().IsEnable(),
				Extra: &selectItemExtra{
					Remark: dict.GetLang(),
					Icon:   dict.GetKey(),
					Color:  dict.GetColor(),
				},
				Value: dict.GetID(),
			}
		}),
	}
}

type SelectDictReply = ListReply[SelectItem]
