package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func NewAlertRecordRepository(d *data.Data) repository.AlertRecord {
	return &alertRecordRepository{Data: d}
}

type alertRecordRepository struct {
	*data.Data
}

func (r *alertRecordRepository) CreateAlertRecord(ctx context.Context, req *bo.CreateAlertRecordBo) (snowflake.ID, error) {
	model := &do.AlertRecord{
		NamespaceUID: contextx.GetNamespace(ctx),
		Source:       req.Source,
		Receiver:     req.Receiver,
		Status:       req.Status,
		Fingerprint:  req.Fingerprint,
		GroupKey:     req.GroupKey,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
		GeneratorURL: req.GeneratorURL,
		Labels:       safety.NewMap(req.Labels),
		Annotations:  safety.NewMap(req.Annotations),
		Raw:          req.Raw,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	if model.Creator == 0 {
		model.WithCreator(model.NamespaceUID)
	}
	if err := r.DB().WithContext(ctx).Create(model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *alertRecordRepository) GetAlertRecord(ctx context.Context, uid snowflake.ID) (*bo.AlertRecordItemBo, error) {
	var model do.AlertRecord
	err := r.DB().WithContext(ctx).Where("namespace_uid = ? AND id = ?", contextx.GetNamespace(ctx), uid).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert record not found")
		}
		return nil, err
	}
	return toAlertRecordItemBo(&model), nil
}

func (r *alertRecordRepository) ListAlertRecord(ctx context.Context, req *bo.ListAlertRecordBo) (*bo.PageResponseBo[*bo.AlertRecordItemBo], error) {
	db := r.DB().WithContext(ctx).Model(&do.AlertRecord{}).Where("namespace_uid = ?", contextx.GetNamespace(ctx))
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("fingerprint LIKE ? OR receiver LIKE ? OR group_key LIKE ?", keyword, keyword, keyword)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Fingerprint != "" {
		db = db.Where("fingerprint = ?", req.Fingerprint)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	req.WithTotal(total)
	var models []*do.AlertRecord
	if err := db.Order("created_at DESC").Offset(req.Offset()).Limit(req.Limit()).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.AlertRecordItemBo, 0, len(models))
	for _, model := range models {
		items = append(items, toAlertRecordItemBo(model))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func toAlertRecordItemBo(model *do.AlertRecord) *bo.AlertRecordItemBo {
	var labels map[string]string
	if model.Labels != nil {
		labels = model.Labels.Map()
	}
	var annotations map[string]string
	if model.Annotations != nil {
		annotations = model.Annotations.Map()
	}
	return &bo.AlertRecordItemBo{
		UID:          model.ID,
		Source:       model.Source,
		Receiver:     model.Receiver,
		Status:       model.Status,
		Fingerprint:  model.Fingerprint,
		GroupKey:     model.GroupKey,
		StartsAt:     model.StartsAt,
		EndsAt:       model.EndsAt,
		GeneratorURL: model.GeneratorURL,
		Labels:       labels,
		Annotations:  annotations,
		Raw:          string(model.Raw),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func NewAlertSubscriptionRepository(d *data.Data) repository.AlertSubscription {
	return &alertSubscriptionRepository{Data: d}
}

type alertSubscriptionRepository struct {
	*data.Data
}

func (r *alertSubscriptionRepository) GetAlertSubscriptionByName(ctx context.Context, name string) (*bo.AlertSubscriptionItemBo, error) {
	var model do.AlertSubscription
	err := r.DB().WithContext(ctx).Where("namespace_uid = ? AND name = ?", contextx.GetNamespace(ctx), name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert subscription not found")
		}
		return nil, err
	}
	return toAlertSubscriptionItemBo(&model), nil
}

func (r *alertSubscriptionRepository) CreateAlertSubscription(ctx context.Context, req *bo.CreateAlertSubscriptionBo) (snowflake.ID, error) {
	model := &do.AlertSubscription{
		NamespaceUID:            contextx.GetNamespace(ctx),
		Name:                    req.Name,
		Remark:                  req.Remark,
		Labels:                  safety.NewMap(req.Labels),
		ExcludeLabels:           safety.NewMap(req.ExcludeLabels),
		RecipientGroupUIDs:      safety.NewSlice(req.RecipientGroupUIDs),
		Members:                 toAlertSubscriptionMembersDO(req.Members),
		DirectMemberEmailConfig: req.DirectMemberEmailConfigUID,
		DirectMemberTemplateUID: req.DirectMemberTemplateUID,
		Status:                  enum.GlobalStatus_ENABLED,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	if model.Creator == 0 {
		model.WithCreator(model.NamespaceUID)
	}
	if err := r.DB().WithContext(ctx).Create(model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *alertSubscriptionRepository) UpdateAlertSubscription(ctx context.Context, req *bo.UpdateAlertSubscriptionBo) error {
	updates := map[string]any{
		"name":                           req.Name,
		"remark":                         req.Remark,
		"labels":                         safety.NewMap(req.Labels),
		"exclude_labels":                 safety.NewMap(req.ExcludeLabels),
		"recipient_group_uids":           safety.NewSlice(req.RecipientGroupUIDs),
		"members":                        toAlertSubscriptionMembersDO(req.Members),
		"direct_member_email_config_uid": req.DirectMemberEmailConfigUID,
		"direct_member_template_uid":     req.DirectMemberTemplateUID,
	}
	result := r.DB().WithContext(ctx).Model(&do.AlertSubscription{}).Where("namespace_uid = ? AND id = ?", contextx.GetNamespace(ctx), req.UID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("alert subscription not found")
	}
	return nil
}

func (r *alertSubscriptionRepository) DeleteAlertSubscription(ctx context.Context, uid snowflake.ID) error {
	result := r.DB().WithContext(ctx).Where("namespace_uid = ? AND id = ?", contextx.GetNamespace(ctx), uid).Delete(&do.AlertSubscription{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("alert subscription not found")
	}
	return nil
}

func (r *alertSubscriptionRepository) GetAlertSubscription(ctx context.Context, uid snowflake.ID) (*bo.AlertSubscriptionItemBo, error) {
	var model do.AlertSubscription
	err := r.DB().WithContext(ctx).Where("namespace_uid = ? AND id = ?", contextx.GetNamespace(ctx), uid).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert subscription not found")
		}
		return nil, err
	}
	return toAlertSubscriptionItemBo(&model), nil
}

func (r *alertSubscriptionRepository) ListAlertSubscription(ctx context.Context, req *bo.ListAlertSubscriptionBo) (*bo.PageResponseBo[*bo.AlertSubscriptionItemBo], error) {
	db := r.DB().WithContext(ctx).Model(&do.AlertSubscription{}).Where("namespace_uid = ?", contextx.GetNamespace(ctx))
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR remark LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	req.WithTotal(total)
	var models []*do.AlertSubscription
	if err := db.Order("created_at DESC").Offset(req.Offset()).Limit(req.Limit()).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.AlertSubscriptionItemBo, 0, len(models))
	for _, model := range models {
		items = append(items, toAlertSubscriptionItemBo(model))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *alertSubscriptionRepository) ListEnabledAlertSubscriptions(ctx context.Context) ([]*bo.AlertSubscriptionItemBo, error) {
	var models []*do.AlertSubscription
	if err := r.DB().WithContext(ctx).Where("namespace_uid = ? AND status = ?", contextx.GetNamespace(ctx), enum.GlobalStatus_ENABLED).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.AlertSubscriptionItemBo, 0, len(models))
	for _, model := range models {
		items = append(items, toAlertSubscriptionItemBo(model))
	}
	return items, nil
}

func (r *alertSubscriptionRepository) UpdateAlertSubscriptionStatus(ctx context.Context, req *bo.UpdateAlertSubscriptionStatusBo) error {
	result := r.DB().WithContext(ctx).Model(&do.AlertSubscription{}).Where("namespace_uid = ? AND id = ?", contextx.GetNamespace(ctx), req.UID).Update("status", req.Status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("alert subscription not found")
	}
	return nil
}

func toAlertSubscriptionItemBo(model *do.AlertSubscription) *bo.AlertSubscriptionItemBo {
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
	members := make([]*bo.AlertSubscriptionMemberBo, 0, len(model.Members))
	for _, item := range model.Members {
		members = append(members, &bo.AlertSubscriptionMemberBo{
			MemberUID: item.MemberUID,
			IsEmail:   item.IsEmail,
			IsSMS:     item.IsSMS,
			IsPhone:   item.IsPhone,
		})
	}
	return &bo.AlertSubscriptionItemBo{
		UID:                        model.ID,
		Name:                       model.Name,
		Remark:                     model.Remark,
		Labels:                     labels,
		ExcludeLabels:              excludeLabels,
		RecipientGroupUIDs:         recipientGroupUIDs,
		Members:                    members,
		DirectMemberEmailConfigUID: model.DirectMemberEmailConfig,
		DirectMemberTemplateUID:    model.DirectMemberTemplateUID,
		Status:                     model.Status,
		CreatedAt:                  model.CreatedAt,
		UpdatedAt:                  model.UpdatedAt,
	}
}

func toAlertSubscriptionMembersDO(items []*bo.AlertSubscriptionMemberBo) do.AlertSubscriptionMembers {
	out := make(do.AlertSubscriptionMembers, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, do.AlertSubscriptionMember{
			MemberUID: item.MemberUID,
			IsEmail:   item.IsEmail,
			IsSMS:     item.IsSMS,
			IsPhone:   item.IsPhone,
		})
	}
	return out
}
