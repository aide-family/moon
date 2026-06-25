package service

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func NewUserService(userBiz *biz.User) *UserService {
	return &UserService{
		userBiz: userBiz,
	}
}

type UserService struct {
	apiv1.UnimplementedUserServer

	userBiz *biz.User
}

func (s *UserService) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.UserItem, error) {
	user, err := s.userBiz.GetUser(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return user.ToAPIV1UserItem(), nil
}

func (s *UserService) ListUser(ctx context.Context, req *apiv1.ListUserRequest) (*apiv1.ListUserReply, error) {
	listBo := bo.NewListUserBo(req)
	page, err := s.userBiz.ListUser(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListUserReply(page), nil
}

func (s *UserService) SelectUser(ctx context.Context, req *apiv1.SelectUserRequest) (*apiv1.SelectUserReply, error) {
	selectBo := bo.NewSelectUserBo(req)
	result, err := s.userBiz.SelectUser(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectUserReply(result), nil
}

func (s *UserService) BanUser(ctx context.Context, req *apiv1.BanUserRequest) (*apiv1.BanUserReply, error) {
	if err := s.userBiz.BanUser(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.BanUserReply{Message: "user has been banned"}, nil
}

func (s *UserService) PermitUser(ctx context.Context, req *apiv1.PermitUserRequest) (*apiv1.PermitUserReply, error) {
	if err := s.userBiz.PermitUser(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.PermitUserReply{Message: "user has been permitted"}, nil
}

func (s *UserService) ValidateUser(ctx context.Context, userUID snowflake.ID) error {
	return s.userBiz.ValidateUser(ctx, userUID)
}

// ListUserFinalMenus 查询系统用户最终菜单集合（9.1）
func (s *UserService) ListUserFinalMenus(ctx context.Context, req *apiv1.ListUserFinalMenusRequest) (*apiv1.ListUserFinalMenusReply, error) {
	listBo := bo.NewListUserFinalMenusBo(req)
	result, err := s.userBiz.ListUserFinalMenus(ctx, listBo)
	if err != nil {
		return nil, err
	}

	if req.AsTree {
		menuItemBos := make([]*bo.SysMenuItemBo, 0, len(result.Items))
		for _, item := range result.Items {
			menuItemBos = append(menuItemBos, &bo.SysMenuItemBo{
				MenuCode:       item.MenuCode,
				ParentMenuCode: item.ParentMenuCode,
				MenuName:       item.MenuName,
				MenuSortOrder:  0,
				MenuStatus:     item.MenuStatus,
			})
		}
		tree := bo.BuildMenuTree(menuItemBos)
		return &apiv1.ListUserFinalMenusReply{
			Tree:      tree,
			MenuCodes: result.MenuCodes,
		}, nil
	}

	apiItems := make([]*apiv1.FinalMenuItem, 0, len(result.Items))
	for _, item := range result.Items {
		apiItems = append(apiItems, &apiv1.FinalMenuItem{
			MenuCode:       item.MenuCode,
			ParentMenuCode: item.ParentMenuCode,
			MenuName:       item.MenuName,
			MenuStatus:     item.MenuStatus,
		})
	}
	return &apiv1.ListUserFinalMenusReply{
		Items:     apiItems,
		MenuCodes: result.MenuCodes,
	}, nil
}

// ListUserFinalFunctions 查询系统用户最终功能集合（9.2）
func (s *UserService) ListUserFinalFunctions(ctx context.Context, req *apiv1.ListUserFinalFunctionsRequest) (*apiv1.ListUserFinalFunctionsReply, error) {
	listBo := bo.NewListUserFinalFunctionsBo(req)
	result, err := s.userBiz.ListUserFinalFunctions(ctx, listBo)
	if err != nil {
		return nil, err
	}

	apiKeys := make([]*apiv1.FinalFunctionItem, 0, len(result.FunctionKeys))
	for _, fk := range result.FunctionKeys {
		apiKeys = append(apiKeys, &apiv1.FinalFunctionItem{
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
			FunctionName: fk.FunctionName,
		})
	}

	apiMap := make(map[string]*apiv1.FinalFunctionList)
	for menuCode, codes := range result.FunctionMap {
		apiMap[menuCode] = &apiv1.FinalFunctionList{FunctionCodes: codes}
	}

	return &apiv1.ListUserFinalFunctionsReply{
		FunctionKeys: apiKeys,
		FunctionMap:  apiMap,
	}, nil
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *UserService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspaceFinalMenus 查询租户成员最终菜单集合（系统+租户合并）
func (s *UserService) ListNmspaceFinalMenus(ctx context.Context, req *apiv1.ListNmspaceFinalMenusRequest) (*apiv1.ListNmspaceFinalMenusReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspaceFinalMenusBo(namespaceUID, req)
	result, err := s.userBiz.ListNmspaceFinalMenus(ctx, listBo)
	if err != nil {
		return nil, err
	}

	if req.AsTree {
		menuItemBos := make([]*bo.SysMenuItemBo, 0, len(result.Items))
		for _, item := range result.Items {
			menuItemBos = append(menuItemBos, &bo.SysMenuItemBo{
				MenuCode:       item.MenuCode,
				ParentMenuCode: item.ParentMenuCode,
				MenuName:       item.MenuName,
				MenuSortOrder:  0,
				MenuStatus:     item.MenuStatus,
			})
		}
		tree := bo.BuildMenuTree(menuItemBos)
		return &apiv1.ListNmspaceFinalMenusReply{
			Tree:      tree,
			MenuCodes: result.MenuCodes,
		}, nil
	}

	apiItems := make([]*apiv1.FinalMenuItem, 0, len(result.Items))
	for _, item := range result.Items {
		apiItems = append(apiItems, &apiv1.FinalMenuItem{
			MenuCode:       item.MenuCode,
			ParentMenuCode: item.ParentMenuCode,
			MenuName:       item.MenuName,
			MenuStatus:     item.MenuStatus,
		})
	}
	return &apiv1.ListNmspaceFinalMenusReply{
		Items:     apiItems,
		MenuCodes: result.MenuCodes,
	}, nil
}

// ListNmspaceFinalFunctions 查询租户成员最终功能集合（系统+租户合并）
func (s *UserService) ListNmspaceFinalFunctions(ctx context.Context, req *apiv1.ListNmspaceFinalFunctionsRequest) (*apiv1.ListNmspaceFinalFunctionsReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspaceFinalFunctionsBo(namespaceUID, req)
	result, err := s.userBiz.ListNmspaceFinalFunctions(ctx, listBo)
	if err != nil {
		return nil, err
	}

	apiKeys := make([]*apiv1.FinalFunctionItem, 0, len(result.FunctionKeys))
	for _, fk := range result.FunctionKeys {
		apiKeys = append(apiKeys, &apiv1.FinalFunctionItem{
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
			FunctionName: fk.FunctionName,
		})
	}

	apiMap := make(map[string]*apiv1.FinalFunctionList)
	for menuCode, codes := range result.FunctionMap {
		apiMap[menuCode] = &apiv1.FinalFunctionList{FunctionCodes: codes}
	}

	return &apiv1.ListNmspaceFinalFunctionsReply{
		FunctionKeys: apiKeys,
		FunctionMap:  apiMap,
	}, nil
}
