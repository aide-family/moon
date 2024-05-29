package resource

import (
	"context"

	pb "github.com/aide-cloud/moon/api/admin/resource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/build"
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
	if err != nil {
		return nil, err
	}
	return &pb.ListMenuReply{
		List: build.NewMenuBuild(menuTree, 0).ToTree(),
	}, nil
}
