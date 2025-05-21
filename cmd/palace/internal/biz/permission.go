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
	menuRepo repository.Menu,
	userRepo repository.User,
	teamRepo repository.Team,
	memberRepo repository.Member,
	logger log.Logger,
) *Permission {
	baseHandler := &basePermissionHandler{}
	// build permission chain
	permissionChain := []PermissionHandler{
		baseHandler.OperationHandler(),
		baseHandler.MenuHandler(menuRepo.GetMenuByOperation),
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
	permissionChain []PermissionHandler // add permission chain
	helper          *log.Helper
}

func (a *Permission) VerifyPermission(ctx context.Context) error {
	pCtx := &PermissionContext{}
	for _, handler := range a.permissionChain {
		skip, err := handler.Handle(ctx, pCtx)
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
	Operation      string
	Menu           do.Menu
	User           do.User
	Team           do.Team
	SystemPosition vobj.Position
	TeamPosition   vobj.Position
	TeamMember     do.TeamMember
}

// PermissionHandler permission handler interface
type PermissionHandler interface {
	Handle(ctx context.Context, pCtx *PermissionContext) (skip bool, err error)
}

// PermissionHandlerFunc permission handler function type
type PermissionHandlerFunc func(ctx context.Context, pCtx *PermissionContext) (skip bool, err error)

func (f PermissionHandlerFunc) Handle(ctx context.Context, pCtx *PermissionContext) (skip bool, err error) {
	return f(ctx, pCtx)
}

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
func (h *basePermissionHandler) OperationHandler() PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		operation, ok := permission.GetOperationByContext(ctx)
		if !ok {
			return true, merr.ErrorBadRequest("operation is invalid")
		}
		pCtx.Operation = operation
		return false, nil
	})
}

// MenuHandler menu check
func (h *basePermissionHandler) MenuHandler(findMenuByOperation FindMenuByOperation) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		menuDo, ok := do.GetMenuDoContext(ctx)
		if !ok {
			var err error
			menuDo, err = findMenuByOperation(ctx, pCtx.Operation)
			if err != nil {
				return true, resourceNotOpen(err)
			}
		}
		pCtx.Menu = menuDo
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
	})
}

type FindUserByID func(ctx context.Context, userID uint32) (do.User, error)
type FindMenuByOperation func(ctx context.Context, operation string) (do.Menu, error)

// UserHandler user check
func (h *basePermissionHandler) UserHandler(findUserByID FindUserByID) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
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
		pCtx.User = userDo
		pCtx.SystemPosition = userDo.GetPosition()
		menuDo := pCtx.Menu
		if menuDo.GetMenuType().IsMenuUser() {
			return true, nil
		}
		return false, nil
	})
}

// SystemAdminCheckHandler system admin check
func (h *basePermissionHandler) SystemAdminCheckHandler() PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		return pCtx.SystemPosition.IsAdminOrSuperAdmin(), nil
	})
}

// SystemRBACHandler system rbac check
func (h *basePermissionHandler) SystemRBACHandler(checkSystemRBAC func(ctx context.Context, user do.User, menu do.Menu) (bool, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		return checkSystemRBAC(ctx, pCtx.User, pCtx.Menu)
	})
}

// TeamIDHandler team id check
func (h *basePermissionHandler) TeamIDHandler(findTeamByID func(ctx context.Context, teamID uint32) (do.Team, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
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
		pCtx.Team = teamItem
		return false, nil
	})
}

// TeamMemberHandler team member check
func (h *basePermissionHandler) TeamMemberHandler(findTeamMemberByUserID func(ctx context.Context, userID uint32) (do.TeamMember, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		member, err := findTeamMemberByUserID(ctx, pCtx.User.GetID())
		if err != nil {
			return true, err
		}
		if !member.GetStatus().IsNormal() {
			return true, merr.ErrorPermissionDenied("team member is invalid [%s]", member.GetStatus())
		}
		if pCtx.Team.GetID() != member.GetTeamID() {
			return true, merr.ErrorPermissionDenied("selected team is invalid")
		}
		pCtx.TeamMember = member
		pCtx.TeamPosition = member.GetPosition()
		return false, nil
	})
}

// TeamAdminCheckHandler team admin check
func (h *basePermissionHandler) TeamAdminCheckHandler() PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		return pCtx.TeamPosition.IsAdminOrSuperAdmin(), nil
	})
}

// TeamRBACHandler team rbac check
func (h *basePermissionHandler) TeamRBACHandler(checkTeamRBAC func(ctx context.Context, member do.TeamMember, menu do.Menu) (bool, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		return checkTeamRBAC(ctx, pCtx.TeamMember, pCtx.Menu)
	})
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
