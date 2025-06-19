package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

// NewPermission create a new permission biz
func NewPermissionBiz(
	cacheRepo repository.Cache,
	userRepo repository.User,
	teamRepo repository.Team,
	memberRepo repository.Member,
	logger log.Logger,
) *Permission {
	baseHandler := &basePermissionHandler{}
	// build permission chain
	permissionChain := []PermissionHandlerFunc{
		baseHandler.OperationHandler(),
		baseHandler.MenuHandler(cacheRepo.GetMenu),
		baseHandler.UserHandler(userRepo.FindByID),
		baseHandler.SystemAdminCheckHandler(),
		baseHandler.SystemRBACHandler(checkSystemRBAC),
		baseHandler.TeamIDHandler(teamRepo.FindByID),
		baseHandler.TeamMemberHandler(memberRepo.FindByUserID),
		baseHandler.TeamAdminCheckHandler(),
		baseHandler.TeamRBACHandler(checkTeamRBAC),
	}
	return &Permission{
		helper:          log.NewHelper(log.With(logger, "module", "biz.permission")),
		permissionChain: permissionChain,
	}
}

type Permission struct {
	permissionChain []PermissionHandlerFunc // add permission chain
	helper          *log.Helper
}

func (a *Permission) VerifyPermission(ctx context.Context) error {
	pCtx := &PermissionContext{Context: ctx}
	for _, handler := range a.permissionChain {
		skip, err := handler(pCtx)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
	}
	return nil
}

// PermissionContext permission check context
type PermissionContext struct {
	context.Context
	Operation      string
	Menu           do.Menu
	User           do.User
	Team           do.Team
	SystemPosition vobj.Position
	TeamPosition   vobj.Position
	TeamMember     do.TeamMember
}

// PermissionHandlerFunc permission handler function type
type PermissionHandlerFunc func(ctx *PermissionContext) (skip bool, err error)

// base permission handler implementation
type basePermissionHandler struct{}

func resourceNotOpen(err error) error {
	if validate.IsNil(err) {
		return nil
	}
	if merr.IsNotFound(err) {
		return merr.ErrorResourceNotOpen("menu")
	}
	return err
}

// OperationHandler operation check
func (h *basePermissionHandler) OperationHandler() PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		operation, ok := permission.GetOperationByContext(ctx)
		if !ok {
			return true, merr.ErrorBadRequest("operation is invalid")
		}
		ctx.Operation = operation
		return false, nil
	}
}

// MenuHandler menu check
func (h *basePermissionHandler) MenuHandler(findMenuByOperation FindMenuByOperation) PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		menuDo, ok := do.GetMenuDoContext(ctx)
		if !ok {
			var err error
			menuDo, err = findMenuByOperation(ctx, ctx.Operation)
			if err != nil {
				return true, resourceNotOpen(err)
			}
		}
		ctx.Menu = menuDo
		if !menuDo.GetProcessType().IsContainsLogin() {
			return true, nil
		}
		if !menuDo.GetStatus().IsEnable() || menuDo.GetDeletedAt() > 0 {
			return false, merr.ErrorPermissionDenied("permission denied")
		}
		if menuDo.GetMenuType().IsMenuNone() {
			return true, nil
		}
		return false, nil
	}
}

type FindUserByID func(ctx context.Context, userID uint32) (do.User, error)
type FindMenuByOperation func(ctx context.Context, operation string) (do.Menu, error)

// UserHandler user check
func (h *basePermissionHandler) UserHandler(findUserByID FindUserByID) PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		userDo, ok := do.GetUserDoContext(ctx)
		if !ok {
			var err error
			userID, ok := permission.GetUserIDByContext(ctx)
			if !ok {
				return true, merr.ErrorBadRequest("user id is invalid")
			}
			userDo, err = findUserByID(ctx, userID)
			if err != nil {
				return true, err
			}
		}

		if validate.IsNil(userDo) {
			return true, merr.ErrorUserForbidden("user is not found")
		}
		if !userDo.GetStatus().IsNormal() {
			return true, merr.ErrorUserForbidden("user is forbidden")
		}
		ctx.User = userDo
		ctx.SystemPosition = userDo.GetPosition()
		menuDo := ctx.Menu
		if menuDo.GetMenuType().IsMenuUser() {
			return true, nil
		}
		return false, nil
	}
}

// SystemAdminCheckHandler system admin check
func (h *basePermissionHandler) SystemAdminCheckHandler() PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		if ctx.SystemPosition.IsAdminOrSuperAdmin() {
			return true, nil
		}
		menuDo := ctx.Menu
		if menuDo.GetMenuType().IsMenuSystem() && menuDo.GetProcessType().IsContainsAdmin() {
			return false, merr.ErrorPermissionDenied("this menu is only available to system administrators")
		}
		return false, nil
	}
}

// SystemRBACHandler system rbac check
func (h *basePermissionHandler) SystemRBACHandler(checkSystemRBAC func(ctx context.Context, user do.User, menu do.Menu) (bool, error)) PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		return checkSystemRBAC(ctx, ctx.User, ctx.Menu)
	}
}

// TeamIDHandler team id check
func (h *basePermissionHandler) TeamIDHandler(findTeamByID func(ctx context.Context, teamID uint32) (do.Team, error)) PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		teamID, ok := permission.GetTeamIDByContext(ctx)
		if !ok {
			return true, merr.ErrorPermissionDenied("please select a team")
		}
		teamItem, err := findTeamByID(ctx, teamID)
		if err != nil {
			return true, err
		}
		if !teamItem.GetStatus().IsNormal() {
			return true, merr.ErrorPermissionDenied("team is invalid")
		}
		ctx.Team = teamItem
		return false, nil
	}
}

// TeamMemberHandler team member check
func (h *basePermissionHandler) TeamMemberHandler(findTeamMemberByUserID func(ctx context.Context, userID uint32) (do.TeamMember, error)) PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		member, err := findTeamMemberByUserID(ctx, ctx.User.GetID())
		if err != nil {
			return true, err
		}
		if !member.GetStatus().IsNormal() {
			return true, merr.ErrorPermissionDenied("team member is invalid [%s]", member.GetStatus())
		}
		if ctx.Team.GetID() != member.GetTeamID() {
			return true, merr.ErrorPermissionDenied("selected team is invalid")
		}
		ctx.TeamMember = member
		ctx.TeamPosition = member.GetPosition()
		return false, nil
	}
}

// TeamAdminCheckHandler team admin check
func (h *basePermissionHandler) TeamAdminCheckHandler() PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		if ctx.TeamPosition.IsAdminOrSuperAdmin() {
			return true, nil
		}
		menuDo := ctx.Menu
		if menuDo.GetMenuType().IsMenuTeam() && menuDo.GetProcessType().IsContainsAdmin() {
			return false, merr.ErrorPermissionDenied("this menu is only available to team administrators")
		}
		return false, nil
	}
}

// TeamRBACHandler team rbac check
func (h *basePermissionHandler) TeamRBACHandler(checkTeamRBAC func(ctx context.Context, member do.TeamMember, menu do.Menu) (bool, error)) PermissionHandlerFunc {
	return func(ctx *PermissionContext) (bool, error) {
		return checkTeamRBAC(ctx, ctx.TeamMember, ctx.Menu)
	}
}

func checkSystemRBAC(_ context.Context, user do.User, menu do.Menu) (bool, error) {
	if !menu.GetMenuType().IsMenuSystem() {
		return false, nil
	}
	resources := make([]uint32, 0, len(user.GetRoles())*10)
	for _, role := range user.GetRoles() {
		if role.GetStatus().IsEnable() {
			menus := slices.MapFilter(role.GetMenus(), func(v do.Menu) (uint32, bool) {
				if validate.IsNil(v) || !v.GetStatus().IsEnable() || v.GetDeletedAt() > 0 {
					return 0, false
				}
				return v.GetID(), true
			})
			resources = append(resources, menus...)
		}
	}
	if slices.Contains(resources, menu.GetID()) {
		return true, nil
	}
	return false, merr.ErrorPermissionDenied("user role is invalid.")
}

func checkTeamRBAC(_ context.Context, member do.TeamMember, menu do.Menu) (bool, error) {
	if !menu.GetMenuType().IsMenuTeam() {
		return false, nil
	}
	roles := member.GetRoles()
	resources := make([]uint32, 0, len(roles)*10)
	for _, role := range roles {
		if role.GetStatus().IsEnable() {
			menus := slices.MapFilter(role.GetMenus(), func(v do.Menu) (uint32, bool) {
				if validate.IsNil(v) || !v.GetStatus().IsEnable() || v.GetDeletedAt() > 0 {
					return 0, false
				}
				return v.GetID(), true
			})
			resources = append(resources, menus...)
		}
	}
	if slices.Contains(resources, menu.GetID()) {
		return true, nil
	}
	return false, merr.ErrorPermissionDenied("team role resourceRepo is invalid.")
}
