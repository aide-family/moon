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
)

// NewPermissionBiz create a new permission biz
func NewPermissionBiz(
	menuRepo repository.Menu,
	userRepo repository.User,
	teamRepo repository.Team,
	memberRepo repository.Member,
	logger log.Logger,
) *PermissionBiz {
	baseHandler := &basePermissionHandler{}
	// build permission chain
	permissionChain := []PermissionHandler{
		baseHandler.UserHandler(userRepo.FindByID),
		baseHandler.OperationHandler(),
		baseHandler.ResourceHandler(menuRepo.GetMenuByOperation),
		baseHandler.AllowCheckHandler(),
		baseHandler.SystemAdminCheckHandler(),
		baseHandler.SystemRBACHandler(checkSystemRBAC),
		baseHandler.TeamIDHandler(teamRepo.FindByID),
		baseHandler.TeamMemberHandler(memberRepo.FindByUserID),
		baseHandler.TeamAdminCheckHandler(),
		baseHandler.TeamRBACHandler(checkTeamRBAC),
	}
	return &PermissionBiz{
		helper:          log.NewHelper(log.With(logger, "module", "biz.permission")),
		permissionChain: permissionChain,
	}
}

type PermissionBiz struct {
	permissionChain []PermissionHandler // add permission chain
	helper          *log.Helper
}

func (a *PermissionBiz) VerifyPermission(ctx context.Context) error {
	pCtx := &PermissionContext{}
	for _, handler := range a.permissionChain {
		stop, err := handler.Handle(ctx, pCtx)
		if err != nil {
			return err
		}
		if stop {
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
	SystemPosition vobj.Role
	TeamPosition   vobj.Role
	TeamMember     do.TeamMember
}

// PermissionHandler permission handler interface
type PermissionHandler interface {
	Handle(ctx context.Context, pCtx *PermissionContext) (stop bool, err error)
}

// PermissionHandlerFunc permission handler function type
type PermissionHandlerFunc func(ctx context.Context, pCtx *PermissionContext) (stop bool, err error)

func (f PermissionHandlerFunc) Handle(ctx context.Context, pCtx *PermissionContext) (bool, error) {
	return f(ctx, pCtx)
}

// base permission handler implementation
type basePermissionHandler struct{}

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

// ResourceHandler resourceRepo check
func (h *basePermissionHandler) ResourceHandler(getMenuByOperation func(ctx context.Context, operation string) (do.Menu, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		menu, err := getMenuByOperation(ctx, pCtx.Operation)
		if err != nil {
			return true, err
		}
		if !menu.GetStatus().IsEnable() {
			return true, merr.ErrorPermissionDenied("permission denied")
		}
		pCtx.Menu = menu
		return false, nil
	})
}

// UserHandler user check
func (h *basePermissionHandler) UserHandler(findUserByID func(ctx context.Context, userID uint32) (do.User, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		userID, ok := permission.GetUserIDByContext(ctx)
		if !ok {
			return true, merr.ErrorBadRequest("user id is invalid")
		}
		user, err := findUserByID(ctx, userID)
		if err != nil {
			return true, err
		}
		if !user.GetStatus().IsNormal() {
			return true, merr.ErrorUserForbidden("user forbidden")
		}
		pCtx.User = user
		pCtx.SystemPosition = user.GetPosition()
		return false, nil
	})
}

// AllowCheckHandler allow check
func (h *basePermissionHandler) AllowCheckHandler() PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		// if pCtx.Menu.GetAllow().IsNone() || pCtx.Menu.GetAllow().IsUser() {
		// 	return true, nil // satisfy condition directly pass
		// }
		return false, nil
	})
}

// SystemAdminCheckHandler system admin check
func (h *basePermissionHandler) SystemAdminCheckHandler() PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		if pCtx.SystemPosition.IsAdminOrSuperAdmin() {
			return true, nil // 管理员直接通过
		}
		return false, nil
	})
}

// SystemRBACHandler system rbac check
func (h *basePermissionHandler) SystemRBACHandler(checkSystemRBAC func(ctx context.Context, user do.User, menu do.Menu) (bool, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		ok, err := checkSystemRBAC(ctx, pCtx.User, pCtx.Menu)
		if err != nil {
			return false, err
		}
		return ok, nil
	})
}

// TeamIDHandler team id check
func (h *basePermissionHandler) TeamIDHandler(findTeamByID func(ctx context.Context, teamID uint32) (do.Team, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		teamID, ok := permission.GetTeamIDByContext(ctx)
		if !ok {
			return true, merr.ErrorPermissionDenied("team id is invalid")
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
			return true, merr.ErrorPermissionDenied("team id is invalid")
		}
		pCtx.TeamMember = member
		pCtx.TeamPosition = member.GetPosition()
		return false, nil
	})
}

// TeamAdminCheckHandler team admin check
func (h *basePermissionHandler) TeamAdminCheckHandler() PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		if pCtx.TeamPosition.IsAdminOrSuperAdmin() {
			return true, nil // team admin directly pass
		}
		return false, nil
	})
}

// TeamRBACHandler team rbac check
func (h *basePermissionHandler) TeamRBACHandler(checkTeamRBAC func(ctx context.Context, member do.TeamMember, menu do.Menu) (bool, error)) PermissionHandler {
	return PermissionHandlerFunc(func(ctx context.Context, pCtx *PermissionContext) (bool, error) {
		ok, err := checkTeamRBAC(ctx, pCtx.TeamMember, pCtx.Menu)
		if err != nil {
			return false, err
		}
		return ok, nil
	})
}

func checkSystemRBAC(_ context.Context, user do.User, menu do.Menu) (bool, error) {
	// if !menu.GetAllow().IsSystemRBAC() {
	// 	return false, nil
	// }
	resources := make([]do.Menu, 0, len(user.GetRoles())*10)
	for _, role := range user.GetRoles() {
		if role.GetStatus().IsEnable() {
			for _, menu := range role.GetMenus() {
				if !menu.GetStatus().IsEnable() {
					continue
				}
				resources = append(resources, menu)
			}
		}
	}
	_, ok := slices.FindByValue(resources, menu.GetID(), func(role do.Menu) uint32 { return role.GetID() })
	if ok {
		return true, nil
	}
	return false, merr.ErrorPermissionDenied("user role resourceRepo is invalid.")
}

func checkTeamRBAC(_ context.Context, member do.TeamMember, menu do.Menu) (bool, error) {
	return true, nil
	// if !menu.GetAllow().IsTeamRBAC() {
	// 	return false, nil
	// }
	// roles := member.GetRoles()
	// resources := make([]do.Menu, 0, len(roles)*10)
	// for _, role := range roles {
	// 	if role.GetStatus().IsEnable() {
	// 		for _, menu := range role.GetMenus() {
	// 			if !menu.GetStatus().IsEnable() {
	// 				continue
	// 			}
	// 			resources = append(resources, menu.GetResources()...)
	// 		}
	// 	}
	// }
	// _, ok := slices.FindByValue(resources, resource.GetID(), func(role do.Resource) uint32 { return role.GetID() })
	// if ok {
	// 	return true, nil
	// }
	// return false, merr.ErrorPermissionDenied("team role resourceRepo is invalid.")
}
