package biz

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

func NewUser(
	userRepo repository.User,
	sysUserRoleRepo repository.SysUserRole,
	sysRoleMenuRepo repository.SysRoleMenu,
	sysRoleFunctionRepo repository.SysRoleFunction,
	sysUserMenuRepo repository.SysUserMenu,
	sysUserFunctionRepo repository.SysUserFunction,
	sysMenuRepo repository.SysMenu,
	sysFunctionRepo repository.SysFunction,
	nmspacesMemberRoleRepo repository.NmspacesMemberRole,
	nmspacesRoleMenuRepo repository.NmspacesRoleMenu,
	nmspacesRoleFunctionRepo repository.NmspacesRoleFunction,
	nmspacesMemberMenuRepo repository.NmspacesMemberMenu,
	nmspacesMemberFunctionRepo repository.NmspacesMemberFunction,
	nmspacesMenuRepo repository.NmspacesMenu,
	helper *klog.Helper,
) *User {
	return &User{
		userRepo:                   userRepo,
		sysUserRoleRepo:            sysUserRoleRepo,
		sysRoleMenuRepo:            sysRoleMenuRepo,
		sysRoleFunctionRepo:        sysRoleFunctionRepo,
		sysUserMenuRepo:            sysUserMenuRepo,
		sysUserFunctionRepo:        sysUserFunctionRepo,
		sysMenuRepo:                sysMenuRepo,
		sysFunctionRepo:            sysFunctionRepo,
		nmspacesMemberRoleRepo:     nmspacesMemberRoleRepo,
		nmspacesRoleMenuRepo:       nmspacesRoleMenuRepo,
		nmspacesRoleFunctionRepo:   nmspacesRoleFunctionRepo,
		nmspacesMemberMenuRepo:     nmspacesMemberMenuRepo,
		nmspacesMemberFunctionRepo: nmspacesMemberFunctionRepo,
		nmspacesMenuRepo:           nmspacesMenuRepo,
		helper:                     klog.NewHelper(klog.With(helper.Logger(), "biz", "user")),
	}
}

type User struct {
	helper                     *klog.Helper
	userRepo                   repository.User
	sysUserRoleRepo            repository.SysUserRole
	sysRoleMenuRepo            repository.SysRoleMenu
	sysRoleFunctionRepo        repository.SysRoleFunction
	sysUserMenuRepo            repository.SysUserMenu
	sysUserFunctionRepo        repository.SysUserFunction
	sysMenuRepo                repository.SysMenu
	sysFunctionRepo            repository.SysFunction
	nmspacesMemberRoleRepo     repository.NmspacesMemberRole
	nmspacesRoleMenuRepo       repository.NmspacesRoleMenu
	nmspacesRoleFunctionRepo   repository.NmspacesRoleFunction
	nmspacesMemberMenuRepo     repository.NmspacesMemberMenu
	nmspacesMemberFunctionRepo repository.NmspacesMemberFunction
	nmspacesMenuRepo           repository.NmspacesMenu
}

func (u *User) GetUserUID(ctx context.Context) (snowflake.ID, error) {
	userUID := contextx.GetUserUID(ctx)
	if userUID <= 0 {
		return 0, merr.ErrorUnauthorized("%s is required", cnst.HTTPHeaderAuthorization)
	}
	return userUID, nil
}

func (u *User) GetUser(ctx context.Context, uid snowflake.ID) (*bo.UserItemBo, error) {
	user, err := u.userRepo.GetUser(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "get user failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get user %s failed", uid).WithCause(err)
	}
	return user, nil
}

func (u *User) ListUser(ctx context.Context, req *bo.ListUserBo) (*bo.PageResponseBo[*bo.UserItemBo], error) {
	page, err := u.userRepo.ListUser(ctx, req)
	if err != nil {
		u.helper.Errorw("msg", "list user failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list user failed").WithCause(err)
	}
	return page, nil
}

func (u *User) SelectUser(ctx context.Context, req *bo.SelectUserBo) (*bo.SelectUserBoResult, error) {
	result, err := u.userRepo.SelectUser(ctx, req)
	if err != nil {
		u.helper.Errorw("msg", "select user failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select user failed").WithCause(err)
	}
	return result, nil
}

func (u *User) BanUser(ctx context.Context, uid snowflake.ID) error {
	if err := u.userRepo.UpdateUserStatus(ctx, uid, enum.UserStatus_BANNED); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "ban user failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("ban user failed").WithCause(err)
	}
	return nil
}

func (u *User) PermitUser(ctx context.Context, uid snowflake.ID) error {
	if err := u.userRepo.UpdateUserStatus(ctx, uid, enum.UserStatus_ACTIVE); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "permit user failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("permit user failed").WithCause(err)
	}
	return nil
}

func (u *User) ChangeEmail(ctx context.Context, uid snowflake.ID, email string) error {
	if err := u.userRepo.UpdateUserEmail(ctx, uid, email); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "change email failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("change email failed").WithCause(err)
	}
	return nil
}

func (u *User) ChangeAvatar(ctx context.Context, uid snowflake.ID, avatar string) error {
	if err := u.userRepo.UpdateUserAvatar(ctx, uid, avatar); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "change avatar failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("change avatar failed").WithCause(err)
	}
	return nil
}

func (u *User) ChangeRemark(ctx context.Context, uid snowflake.ID, remark string) error {
	if err := u.userRepo.UpdateUserRemark(ctx, uid, remark); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "change remark failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("change remark failed").WithCause(err)
	}
	return nil
}

func (u *User) ValidateUser(ctx context.Context, userUID snowflake.ID) error {
	user, err := u.userRepo.GetUser(ctx, userUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorUnauthorized("user is not allowed to access this resource")
		}
		u.helper.Errorw("msg", "validate user failed", "error", err, "userUID", userUID)
		return merr.ErrorInternalServer("validate user failed").WithCause(err)
	}
	if user.Status != enum.UserStatus_ACTIVE {
		return merr.ErrorUnauthorized("user is not active")
	}
	return nil
}

// ListUserFinalMenus 9.1 查询系统用户最终菜单集合
func (u *User) ListUserFinalMenus(ctx context.Context, req *bo.ListUserFinalMenusBo) (*bo.ListUserFinalMenusResultBo, error) {
	// 1. 查询用户通过角色拥有的菜单: user_roles → role_menus → menus
	roleCodes, err := u.sysUserRoleRepo.ListRoleCodesByUserUID(ctx, req.UserUID)
	if err != nil {
		u.helper.Errorw("msg", "list user role codes failed", "error", err)
		return nil, merr.ErrorInternalServer("list user role codes failed").WithCause(err)
	}

	roleMenuCodes, err := u.sysRoleMenuRepo.ListMenuCodesByRoleCodes(ctx, roleCodes)
	if err != nil {
		u.helper.Errorw("msg", "list role menu codes failed", "error", err)
		return nil, merr.ErrorInternalServer("list role menu codes failed").WithCause(err)
	}

	// 2. 查询用户直授菜单: user_menus → menus
	directItems, err := u.sysUserMenuRepo.ListUserMenu(ctx, &bo.ListUserMenuBo{UserUID: req.UserUID})
	if err != nil {
		u.helper.Errorw("msg", "list user direct menus failed", "error", err)
		return nil, merr.ErrorInternalServer("list user direct menus failed").WithCause(err)
	}

	// 3. 合并去重
	menuCodeSet := make(map[string]struct{})
	for _, mc := range roleMenuCodes {
		menuCodeSet[mc] = struct{}{}
	}
	for _, item := range directItems {
		menuCodeSet[item.MenuCode] = struct{}{}
	}

	// 4. 查询所有合并后的菜单详情并过滤禁用/已删除
	allMenuCodes := make([]string, 0, len(menuCodeSet))
	for mc := range menuCodeSet {
		allMenuCodes = append(allMenuCodes, mc)
	}

	var items []*bo.UserFinalMenuItemBo
	for _, mc := range allMenuCodes {
		menu, err := u.sysMenuRepo.GetSysMenu(ctx, mc)
		if err != nil {
			u.helper.Errorw("msg", "get menu failed", "error", err, "menuCode", mc)
			continue
		}
		if menu == nil {
			continue
		}
		items = append(items, &bo.UserFinalMenuItemBo{
			MenuCode:       menu.MenuCode,
			ParentMenuCode: menu.ParentMenuCode,
			MenuName:       menu.MenuName,
			MenuStatus:     menu.MenuStatus,
		})
	}

	return &bo.ListUserFinalMenusResultBo{
		Items:     items,
		MenuCodes: allMenuCodes,
	}, nil
}

// ListUserFinalFunctions 9.2 查询系统用户最终功能集合
func (u *User) ListUserFinalFunctions(ctx context.Context, req *bo.ListUserFinalFunctionsBo) (*bo.ListUserFinalFunctionsResultBo, error) {
	// 1. 查询用户通过角色拥有的功能: user_roles → role_functions → functions
	roleCodes, err := u.sysUserRoleRepo.ListRoleCodesByUserUID(ctx, req.UserUID)
	if err != nil {
		u.helper.Errorw("msg", "list user role codes failed", "error", err)
		return nil, merr.ErrorInternalServer("list user role codes failed").WithCause(err)
	}

	roleFunctions, err := u.sysRoleFunctionRepo.ListFunctionsByRoleCodes(ctx, roleCodes)
	if err != nil {
		u.helper.Errorw("msg", "list role functions failed", "error", err)
		return nil, merr.ErrorInternalServer("list role functions failed").WithCause(err)
	}

	// 2. 查询用户直授功能: user_functions → functions
	directItems, _, err := u.sysUserFunctionRepo.ListUserFunction(ctx, &bo.ListUserFunctionBo{
		UserUID:       req.UserUID,
		MenuCode:      req.MenuCode,
		PageRequestBo: &bo.PageRequestBo{Page: 0, PageSize: 0}, // 不分页
	})
	if err != nil {
		u.helper.Errorw("msg", "list user direct functions failed", "error", err)
		return nil, merr.ErrorInternalServer("list user direct functions failed").WithCause(err)
	}

	// 3. 合并去重 (menuCode, functionCode)
	type funcKey struct {
		menuCode     string
		functionCode string
	}
	seen := make(map[funcKey]struct{})
	var functionKeys []*bo.UserFinalFunctionItemBo

	for _, rf := range roleFunctions {
		if req.MenuCode != "" && rf.MenuCode != req.MenuCode {
			continue
		}
		key := funcKey{rf.MenuCode, rf.FunctionCode}
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			functionKeys = append(functionKeys, &bo.UserFinalFunctionItemBo{
				MenuCode:     rf.MenuCode,
				FunctionCode: rf.FunctionCode,
				FunctionName: rf.FunctionName,
			})
		}
	}

	for _, df := range directItems {
		key := funcKey{df.MenuCode, df.FunctionCode}
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			functionKeys = append(functionKeys, &bo.UserFinalFunctionItemBo{
				MenuCode:     df.MenuCode,
				FunctionCode: df.FunctionCode,
				FunctionName: df.FunctionName,
			})
		}
	}

	// 4. 按 menuCode 分组
	functionMap := make(map[string][]string)
	for _, fk := range functionKeys {
		functionMap[fk.MenuCode] = append(functionMap[fk.MenuCode], fk.FunctionCode)
	}

	return &bo.ListUserFinalFunctionsResultBo{
		FunctionKeys: functionKeys,
		FunctionMap:  functionMap,
	}, nil
}

// ListNmspaceFinalMenus 查询租户成员最终菜单集合（角色菜单 ∪ 直授菜单，按 menuCode 去重）
func (u *User) ListNmspaceFinalMenus(ctx context.Context, req *bo.ListNmspaceFinalMenusBo) (*bo.ListNmspaceFinalMenusResultBo, error) {
	// 1. 查询成员通过角色拥有的菜单: member_roles → role_menus
	roleCodes, err := u.nmspacesMemberRoleRepo.ListRoleCodesByNamespaceAndMemberUID(ctx, req.NamespaceUID, req.MemberUID)
	if err != nil {
		u.helper.Errorw("msg", "list nmspaces member role codes failed", "error", err)
		return nil, merr.ErrorInternalServer("list nmspaces member role codes failed").WithCause(err)
	}

	roleMenuCodes, err := u.nmspacesRoleMenuRepo.ListMenuCodesByNamespaceAndRoleCodes(ctx, req.NamespaceUID, roleCodes)
	if err != nil {
		u.helper.Errorw("msg", "list nmspaces role menu codes failed", "error", err)
		return nil, merr.ErrorInternalServer("list nmspaces role menu codes failed").WithCause(err)
	}

	// 2. 查询成员直授菜单: member_menus
	directItems, err := u.nmspacesMemberMenuRepo.ListNmspacesMemberMenu(ctx, &bo.ListNmspacesMemberMenuBo{
		NamespaceUID: req.NamespaceUID,
		MemberUID:    req.MemberUID,
	})
	if err != nil {
		u.helper.Errorw("msg", "list nmspaces member direct menus failed", "error", err)
		return nil, merr.ErrorInternalServer("list nmspaces member direct menus failed").WithCause(err)
	}

	// 3. 合并去重
	menuCodeSet := make(map[string]struct{})
	for _, mc := range roleMenuCodes {
		menuCodeSet[mc] = struct{}{}
	}
	for _, item := range directItems {
		menuCodeSet[item.MenuCode] = struct{}{}
	}

	// 4. 查询所有合并后的菜单详情
	allMenuCodes := make([]string, 0, len(menuCodeSet))
	for mc := range menuCodeSet {
		allMenuCodes = append(allMenuCodes, mc)
	}

	var items []*bo.NmspaceFinalMenuItemBo
	for _, mc := range allMenuCodes {
		menu, err := u.nmspacesMenuRepo.GetNmspacesMenu(ctx, req.NamespaceUID, mc)
		if err != nil {
			u.helper.Errorw("msg", "get nmspaces menu failed", "error", err, "menuCode", mc)
			continue
		}
		if menu == nil {
			continue
		}
		items = append(items, &bo.NmspaceFinalMenuItemBo{
			MenuCode:       menu.MenuCode,
			ParentMenuCode: menu.ParentMenuCode,
			MenuName:       menu.MenuName,
			MenuStatus:     menu.MenuStatus,
		})
	}

	return &bo.ListNmspaceFinalMenusResultBo{
		Items:     items,
		MenuCodes: allMenuCodes,
	}, nil
}

// ListNmspaceFinalFunctions 查询租户成员最终功能集合（角色功能 ∪ 直授功能，按 menuCode+functionCode 去重）
func (u *User) ListNmspaceFinalFunctions(ctx context.Context, req *bo.ListNmspaceFinalFunctionsBo) (*bo.ListNmspaceFinalFunctionsResultBo, error) {
	// 1. 查询成员通过角色拥有的功能: member_roles → role_functions
	roleCodes, err := u.nmspacesMemberRoleRepo.ListRoleCodesByNamespaceAndMemberUID(ctx, req.NamespaceUID, req.MemberUID)
	if err != nil {
		u.helper.Errorw("msg", "list nmspaces member role codes failed", "error", err)
		return nil, merr.ErrorInternalServer("list nmspaces member role codes failed").WithCause(err)
	}

	roleFunctions, err := u.nmspacesRoleFunctionRepo.ListFunctionsByNamespaceAndRoleCodes(ctx, req.NamespaceUID, roleCodes)
	if err != nil {
		u.helper.Errorw("msg", "list nmspaces role functions failed", "error", err)
		return nil, merr.ErrorInternalServer("list nmspaces role functions failed").WithCause(err)
	}

	// 2. 查询成员直授功能: member_functions
	directItems, _, err := u.nmspacesMemberFunctionRepo.ListNmspacesMemberFunction(ctx, &bo.ListNmspacesMemberFunctionBo{
		NamespaceUID:  req.NamespaceUID,
		MemberUID:     req.MemberUID,
		MenuCode:      req.MenuCode,
		PageRequestBo: &bo.PageRequestBo{Page: 0, PageSize: 0},
	})
	if err != nil {
		u.helper.Errorw("msg", "list nmspaces member direct functions failed", "error", err)
		return nil, merr.ErrorInternalServer("list nmspaces member direct functions failed").WithCause(err)
	}

	// 3. 合并去重 (menuCode, functionCode)
	type funcKey struct {
		menuCode     string
		functionCode string
	}
	seen := make(map[funcKey]struct{})
	var functionKeys []*bo.NmspaceFinalFunctionItemBo

	for _, rf := range roleFunctions {
		if req.MenuCode != "" && rf.MenuCode != req.MenuCode {
			continue
		}
		key := funcKey{rf.MenuCode, rf.FunctionCode}
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			functionKeys = append(functionKeys, &bo.NmspaceFinalFunctionItemBo{
				MenuCode:     rf.MenuCode,
				FunctionCode: rf.FunctionCode,
				FunctionName: rf.FunctionName,
			})
		}
	}

	for _, df := range directItems {
		key := funcKey{df.MenuCode, df.FunctionCode}
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			functionKeys = append(functionKeys, &bo.NmspaceFinalFunctionItemBo{
				MenuCode:     df.MenuCode,
				FunctionCode: df.FunctionCode,
				FunctionName: df.FunctionName,
			})
		}
	}

	// 4. 按 menuCode 分组
	functionMap := make(map[string][]string)
	for _, fk := range functionKeys {
		functionMap[fk.MenuCode] = append(functionMap[fk.MenuCode], fk.FunctionCode)
	}

	return &bo.ListNmspaceFinalFunctionsResultBo{
		FunctionKeys: functionKeys,
		FunctionMap:  functionMap,
	}, nil
}
