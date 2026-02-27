package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
	"github.com/aide-family/goddess/internal/data/impl/query"
)

func NewMemberRepository(d *data.Data) repository.Member {
	return &memberRepository{Data: d}
}

type memberRepository struct {
	*data.Data
}

func (m *memberRepository) CreateMember(ctx context.Context, req *bo.CreateMemberBo) error {
	member := &do.Member{
		Creator:      req.Creator,
		NamespaceUID: req.NamespaceUID,
		UserUID:      req.UserUID,
		Name:         req.Name,
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
	}
	return query.Member.WithContext(ctx).Create(member)
}

func (m *memberRepository) DeleteMember(ctx context.Context, uid snowflake.ID) error {
	mutation := query.Member
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).Delete()
	return err
}

func (m *memberRepository) GetMember(ctx context.Context, uid snowflake.ID) (*bo.MemberItemBo, error) {
	mutation := query.Member
	member, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("member %s not found", uid)
		}
		return nil, err
	}
	email := m.getUserEmail(ctx, member.UserUID)
	return convert.MemberToBo(member, email), nil
}

func (m *memberRepository) GetMemberByNamespaceAndUser(ctx context.Context, namespaceUID, userUID snowflake.ID) (*bo.MemberItemBo, error) {
	mutation := query.Member
	member, err := mutation.WithContext(ctx).
		Where(mutation.NamespaceUID.Eq(namespaceUID.Int64())).
		Where(mutation.UserUID.Eq(userUID.Int64())).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("member not found")
		}
		return nil, err
	}
	email := m.getUserEmail(ctx, member.UserUID)
	return convert.MemberToBo(member, email), nil
}

func (m *memberRepository) getUserEmail(ctx context.Context, userUID snowflake.ID) string {
	user, err := query.User.WithContext(ctx).Where(query.User.UID.Eq(userUID.Int64())).First()
	if err != nil {
		return ""
	}
	return user.Email
}

func (m *memberRepository) getEmailsByUserUIDs(ctx context.Context, userUIDs []int64) map[int64]string {
	if len(userUIDs) == 0 {
		return nil
	}
	users, err := query.User.WithContext(ctx).Where(query.User.UID.In(userUIDs...)).Find()
	if err != nil {
		return nil
	}
	emails := make(map[int64]string, len(users))
	for _, u := range users {
		emails[u.UID.Int64()] = u.Email
	}
	return emails
}

func (m *memberRepository) ListMember(ctx context.Context, req *bo.ListMemberBo) (*bo.PageResponseBo[*bo.MemberItemBo], error) {
	namespaceUID := contextx.GetNamespace(ctx)
	if namespaceUID <= 0 {
		return nil, merr.ErrorInvalidArgument("namespace is required")
	}

	mutation := query.Member
	wrappers := mutation.WithContext(ctx).Where(mutation.NamespaceUID.Eq(namespaceUID.Int64()))
	if req.UserUID > 0 {
		wrappers = wrappers.Where(mutation.UserUID.Eq(req.UserUID.Int64()))
	}
	if len(req.UIDs) > 0 {
		wrappers = wrappers.Where(mutation.UID.In(req.UIDs...))
	}
	if req.Status > enum.MemberStatus_MemberStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(int32(req.Status)))
	}
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.Name.Like(keyword)).
			Or(mutation.Nickname.Like(keyword))
	}
	if strutil.IsNotEmpty(req.Email) {
		users, _ := query.User.WithContext(ctx).
			Where(query.User.Email.Like("%" + req.Email + "%")).
			Find()
		if len(users) > 0 {
			ids := make([]int64, 0, len(users))
			for _, u := range users {
				ids = append(ids, u.UID.Int64())
			}
			wrappers = wrappers.Where(mutation.UserUID.In(ids...))
		}
	}
	if req.Page > 0 && req.PageSize > 0 {
		total, err := wrappers.Count()
		if err != nil {
			return nil, merr.ErrorInternalServer("list member failed: %v", err)
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}
	members, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list member failed: %v", err)
	}
	userUIDs := make([]int64, 0, len(members))
	for _, mb := range members {
		userUIDs = append(userUIDs, mb.UserUID.Int64())
	}
	emails := m.getEmailsByUserUIDs(ctx, userUIDs)
	items := make([]*bo.MemberItemBo, 0, len(members))
	for _, member := range members {
		items = append(items, convert.MemberToBo(member, emails[member.UserUID.Int64()]))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (m *memberRepository) SelectMember(ctx context.Context, req *bo.SelectMemberBo) (*bo.SelectMemberBoResult, error) {
	namespaceUID := contextx.GetNamespace(ctx)
	if namespaceUID <= 0 {
		return nil, merr.ErrorInvalidArgument("namespace is required")
	}

	mutation := query.Member
	wrappers := mutation.WithContext(ctx).Where(mutation.NamespaceUID.Eq(namespaceUID.Int64()))
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.Name.Like(keyword)).
			Or(mutation.Nickname.Like(keyword))
	}
	if req.Status > enum.MemberStatus_MemberStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(int32(req.Status)))
	}
	if req.LastUID > 0 {
		wrappers = wrappers.Where(mutation.UID.Lt(req.LastUID.Int64()))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("select member failed: %v", err)
	}
	members, err := wrappers.Limit(int(req.Limit)).Order(mutation.UID.Desc()).Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("select member failed: %v", err)
	}
	userUIDs := make([]int64, 0, len(members))
	for _, mb := range members {
		userUIDs = append(userUIDs, mb.UserUID.Int64())
	}
	emails := m.getEmailsByUserUIDs(ctx, userUIDs)
	items := make([]*bo.SelectMemberItemBo, 0, len(members))
	for _, member := range members {
		label := member.Nickname
		if label == "" {
			label = member.Name
		}
		items = append(items, &bo.SelectMemberItemBo{
			Value:    member.UID,
			Label:    label,
			Disabled: member.Status != enum.MemberStatus_JOINED || member.DeletedAt.Valid,
			Tooltip:  emails[member.UserUID.Int64()],
		})
	}
	var lastUID snowflake.ID
	if len(members) > 0 {
		lastUID = members[len(members)-1].UID
	}
	return &bo.SelectMemberBoResult{
		Items:   items,
		Total:   total,
		LastUID: lastUID,
		HasMore: len(items) >= int(req.Limit),
	}, nil
}

func (m *memberRepository) UpdateMemberStatus(ctx context.Context, uid snowflake.ID, status int32) error {
	mutation := query.Member
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).Update(mutation.Status, status)
	return err
}

func (m *memberRepository) GetNamespaceUIDsByUserUID(ctx context.Context, userUID snowflake.ID) ([]snowflake.ID, error) {
	mutation := query.Member
	members, err := mutation.WithContext(ctx).
		Where(mutation.UserUID.Eq(userUID.Int64())).
		Select(mutation.NamespaceUID).
		Find()
	if err != nil {
		return nil, err
	}
	uids := make([]snowflake.ID, 0, len(members))
	seen := make(map[int64]bool)
	for _, mb := range members {
		if !seen[mb.NamespaceUID.Int64()] {
			seen[mb.NamespaceUID.Int64()] = true
			uids = append(uids, mb.NamespaceUID)
		}
	}
	return uids, nil
}
