// Package gormimpl is the implementation of the gorm repository for the member service.
package gormimpl

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	memberv1 "github.com/aide-family/magicbox/domain/member/v1"
	"github.com/aide-family/magicbox/domain/member/v1/gormimpl/query"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
)

func init() {
	memberv1.RegisterMemberV1Factory(config.DomainConfig_GORM, NewGormRepository)
}

func NewGormRepository(c *config.DomainConfig) (apiv1.MemberServer, func() error, error) {
	ormConfig := &config.ORMConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), ormConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal orm config failed: %v", err)
		}
	}
	db, close, err := connect.NewDB(ormConfig)
	if err != nil {
		return nil, nil, err
	}
	query.SetDefault(db)

	return &gormRepository{repoConfig: c, db: db}, close, nil
}

type gormRepository struct {
	apiv1.UnimplementedMemberServer
	repoConfig *config.DomainConfig
	db         *gorm.DB
}

// GetMember implements [v1.MemberServer].
func (g *gormRepository) GetMember(ctx context.Context, req *apiv1.GetMemberRequest) (*apiv1.MemberItem, error) {
	m, err := query.Member.WithContext(ctx).Where(query.Member.ID.Eq(req.GetUid())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("member not found")
		}
		return nil, merr.ErrorInternalServer("get member failed: %v", err)
	}
	return convertMemberItem(m), nil
}

// ListMember implements [v1.MemberServer].
func (g *gormRepository) ListMember(ctx context.Context, req *apiv1.ListMemberRequest) (*apiv1.ListMemberReply, error) {
	mut := query.Member
	wrappers := mut.WithContext(ctx)
	if strutil.IsNotEmpty(req.GetKeyword()) {
		kw := "%" + req.GetKeyword() + "%"
		wrappers = wrappers.Or(mut.Name.Like(kw), mut.Nickname.Like(kw), mut.Remark.Like(kw))
	}
	if req.GetStatus() > enum.MemberStatus_MemberStatus_UNKNOWN {
		wrappers = wrappers.Where(mut.Status.Eq(int32(req.GetStatus())))
	}
	if req.GetUserUID() > 0 {
		wrappers = wrappers.Where(mut.UserUID.Eq(req.GetUserUID()))
	}
	if len(req.GetUids()) > 0 {
		wrappers = wrappers.Where(mut.UserUID.In(req.GetUids()...))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("count member failed: %v", err)
	}
	if req.GetPage() > 0 && req.GetPageSize() > 0 {
		wrappers = wrappers.Limit(int(req.GetPageSize())).Offset(int(req.GetPageSize() * (req.GetPage() - 1)))
	}
	wrappers = wrappers.Order(mut.ID.Desc())
	list, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list member failed: %v", err)
	}
	items := make([]*apiv1.MemberItem, 0, len(list))
	for _, m := range list {
		items = append(items, convertMemberItem(m))
	}
	return &apiv1.ListMemberReply{
		Items:    items,
		Total:    total,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// SelectMember implements [v1.MemberServer].
func (g *gormRepository) SelectMember(ctx context.Context, req *apiv1.SelectMemberRequest) (*apiv1.SelectMemberReply, error) {
	mut := query.Member
	wrappers := mut.WithContext(ctx)
	if strutil.IsNotEmpty(req.GetKeyword()) {
		kw := "%" + req.GetKeyword() + "%"
		wrappers = wrappers.Or(mut.Name.Like(kw), mut.Nickname.Like(kw), mut.Remark.Like(kw))
	}
	if req.GetStatus() > enum.MemberStatus_MemberStatus_UNKNOWN {
		wrappers = wrappers.Where(mut.Status.Eq(int32(req.GetStatus())))
	}
	if req.GetLastUID() > 0 {
		wrappers = wrappers.Where(mut.ID.Lt(req.GetLastUID()))
	}
	if len(req.GetUids()) > 0 {
		wrappers = wrappers.Where(mut.UserUID.In(req.GetUids()...))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("count member failed: %v", err)
	}
	wrappers = wrappers.Limit(int(req.GetLimit())).Order(mut.ID.Desc())
	list, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("select member failed: %v", err)
	}
	items := make([]*apiv1.SelectMemberItem, 0, len(list))
	var lastUID int64
	for _, m := range list {
		items = append(items, convertMemberSelectItem(m))
		lastUID = m.ID.Int64()
	}
	return &apiv1.SelectMemberReply{
		Items:   items,
		Total:   total,
		LastUID: lastUID,
		HasMore: len(list) == int(req.GetLimit()),
	}, nil
}
