package resource

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/util/types"
)

type MenuService struct {
	pb.UnimplementedMenuServer

	menuBiz *biz.MenuBiz
}

func NewMenuService(menuBiz *biz.MenuBiz) *MenuService {
	return &MenuService{
		menuBiz: menuBiz,
	}
}

func (s *MenuService) ListMenu(ctx context.Context, _ *pb.ListMenuRequest) (*pb.ListMenuReply, error) {
	menuTree, err := s.menuBiz.MenuList(ctx)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListMenuReply{
		List: build.NewMenuBuild(menuTree, 0).ToTree(),
	}, nil
}
