package convert

import (
	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToAlertSubscriptionItemBo(model *do.AlertSubscription) *bo.AlertSubscriptionItemBo {
	if model == nil {
		return nil
	}
	var labels map[string]string
	if model.Labels != nil {
		labels = model.Labels.Map()
	}
	var excludeLabels map[string]string
	if model.ExcludeLabels != nil {
		excludeLabels = model.ExcludeLabels.Map()
	}
	var recipientGroupUIDs []int64
	if model.RecipientGroupUIDs != nil {
		recipientGroupUIDs = model.RecipientGroupUIDs.List()
	}
	return &bo.AlertSubscriptionItemBo{
		UID:                        model.ID,
		Name:                       model.Name,
		Remark:                     model.Remark,
		Labels:                     labels,
		ExcludeLabels:              excludeLabels,
		RecipientGroupUIDs:         recipientGroupUIDs,
		Members:                    ToNotificationMembersBo(model.Members),
		DirectMemberEmailConfigUID: model.DirectMemberEmailConfig,
		DirectMemberTemplateUID:    model.DirectMemberTemplateUID,
		Status:                     model.Status,
		CreatedAt:                  model.CreatedAt,
		UpdatedAt:                  model.UpdatedAt,
	}
}

func ToAlertSubscriptionDetailBo(
	model *do.AlertSubscription,
	recipientGroups []*do.RecipientGroup,
	emailConfig *do.EmailConfig,
	template *do.Template,
) *bo.AlertSubscriptionDetailBo {
	item := ToAlertSubscriptionItemBo(model)
	if item == nil {
		return nil
	}
	detail := &bo.AlertSubscriptionDetailBo{
		AlertSubscriptionItemBo: *item,
	}
	if len(recipientGroups) > 0 {
		groupMap := make(map[int64]*do.RecipientGroup, len(recipientGroups))
		for _, group := range recipientGroups {
			if group != nil {
				groupMap[group.ID.Int64()] = group
			}
		}
		orderedUIDs := recipientGroupUIDsFromModel(model)
		detail.RecipientGroups = make([]*bo.RecipientGroupItemBo, 0, len(orderedUIDs))
		for _, uid := range orderedUIDs {
			if group, ok := groupMap[uid]; ok {
				detail.RecipientGroups = append(detail.RecipientGroups, ToRecipientGroupItemBo(group))
			}
		}
	}
	if emailConfig != nil {
		detail.DirectMemberEmailConfig = ToEmailConfigBO(emailConfig)
	}
	if template != nil {
		detail.DirectMemberTemplate = ToTemplateItemBo(template)
	}
	return detail
}

func recipientGroupUIDsFromModel(model *do.AlertSubscription) []int64 {
	if model == nil || model.RecipientGroupUIDs == nil {
		return nil
	}
	return model.RecipientGroupUIDs.List()
}

func ToAlertSubscriptionRecipientGroupUIDs(model *do.AlertSubscription) []int64 {
	return recipientGroupUIDsFromModel(model)
}
