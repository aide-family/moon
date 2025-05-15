package impl

import (
	"context"
	_ "embed"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/conf"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/cmd/palace/internal/data/impl/build"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/password"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/template"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewUserRepo(bc *conf.Bootstrap, data *data.Data, logger log.Logger) repository.User {
	return &userRepoImpl{
		bc:     bc,
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.user")),
	}
}

type userRepoImpl struct {
	bc *conf.Bootstrap
	*data.Data
	helper *log.Helper
}

func (u *userRepoImpl) Get(ctx context.Context, id uint32) (do.User, error) {
	return u.FindByID(ctx, id)
}

func (u *userRepoImpl) UpdateUserRoles(ctx context.Context, req bo.UpdateUserRoles) error {
	userMutation := getMainQuery(ctx, u).User
	userDo := &system.User{
		BaseModel: do.BaseModel{
			ID: req.GetUser().GetID(),
		},
	}
	roles := slices.MapFilter(req.GetRoles(), func(role do.Role) (*system.Role, bool) {
		if validate.IsNil(role) || role.GetID() <= 0 {
			return nil, false
		}
		return &system.Role{
			CreatorModel: do.CreatorModel{
				BaseModel: do.BaseModel{
					ID: role.GetID(),
				},
			},
		}, true
	})
	userMutation.WithContext(ctx)
	rolesAssociation := userMutation.Roles.WithContext(ctx).Model(userDo)
	if len(roles) == 0 {
		return rolesAssociation.Clear()
	}
	return rolesAssociation.Replace(roles...)
}

func (u *userRepoImpl) Find(ctx context.Context, ids []uint32) ([]do.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	mutation := getMainQuery(ctx, u)
	user := mutation.User
	userDo, err := user.WithContext(ctx).Where(user.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	users := slices.Map(userDo, func(user *system.User) do.User { return user })
	return users, nil
}

func (u *userRepoImpl) AppendTeam(ctx context.Context, team do.Team) error {
	mutation := getMainQuery(ctx, u)
	userMutation := mutation.User
	userDo := &system.User{
		BaseModel: do.BaseModel{
			ID: team.GetLeaderID(),
		},
	}
	teamDo := build.ToTeam(ctx, team)
	userDo.WithContext(ctx)
	return userMutation.Teams.WithContext(ctx).Model(userDo).Append(teamDo)
}

func (u *userRepoImpl) CreateUserWithOAuthUser(ctx context.Context, user bo.IOAuthUser, sendEmailFunc bo.SendEmailFun) (userDo do.User, err error) {
	userDo, err = u.FindByEmail(ctx, user.GetEmail())
	if err == nil {
		return userDo, nil
	}
	if !merr.IsUserNotFound(err) {
		return nil, err
	}
	userDo = &system.User{
		BaseModel: do.BaseModel{},
		Username:  user.GetUsername(),
		Nickname:  user.GetNickname(),
		Password:  "",
		Email:     user.GetEmail(),
		Phone:     "",
		Remark:    user.GetRemark(),
		Avatar:    user.GetAvatar(),
		Salt:      "",
		Gender:    vobj.GenderUnknown,
		Position:  vobj.RoleUser,
		Status:    vobj.UserStatusNormal,
		Roles:     nil,
	}
	userDo.WithContext(ctx)
	return u.Create(ctx, userDo, sendEmailFunc)
}

func (u *userRepoImpl) Create(ctx context.Context, user do.User, sendEmailFunc bo.SendEmailFun) (do.User, error) {
	pass := password.New(password.GenerateRandomPassword(8))
	enValue, err := pass.EnValue()
	if err != nil {
		return nil, err
	}
	userDo := &system.User{
		Username: user.GetUsername(),
		Nickname: user.GetNickname(),
		Password: enValue,
		Email:    user.GetEmail(),
		Phone:    user.GetPhone(),
		Remark:   user.GetRemark(),
		Avatar:   user.GetAvatar(),
		Salt:     pass.Salt(),
		Gender:   user.GetGender(),
		Position: user.GetPosition(),
		Status:   user.GetStatus(),
	}
	userDo.WithContext(ctx)
	mutation := getMainQuery(ctx, u)
	userMutation := mutation.User
	if err = userMutation.WithContext(ctx).Create(userDo); err != nil {
		return nil, err
	}
	if err = u.sendUserPassword(ctx, user, pass.PValue(), sendEmailFunc); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userRepoImpl) FindByID(ctx context.Context, userID uint32) (do.User, error) {
	mutation := getMainQuery(ctx, u)
	userQuery := mutation.User
	user, err := userQuery.WithContext(ctx).Where(userQuery.ID.Eq(userID)).Preload(userQuery.Roles.Menus.RelationField).First()
	if err != nil {
		return nil, userNotFound(err)
	}
	return user, nil
}

func (u *userRepoImpl) FindByEmail(ctx context.Context, email string) (do.User, error) {
	mutation := getMainQuery(ctx, u)
	userQuery := mutation.User
	user, err := userQuery.WithContext(ctx).Where(userQuery.Email.Eq(email)).First()
	if err != nil {
		return nil, userNotFound(err)
	}
	return user, nil
}

func (u *userRepoImpl) SetEmail(ctx context.Context, user do.User, sendEmailFunc bo.SendEmailFun) (do.User, error) {
	userMutation := getMainQuery(ctx, u).User
	wrapper := []gen.Condition{
		userMutation.ID.Eq(user.GetID()),
		userMutation.Email.Eq(""),
	}
	pass := password.New(password.GenerateRandomPassword(8))
	enValue, err := pass.EnValue()
	if err != nil {
		return nil, err
	}
	mutations := []field.AssignExpr{
		userMutation.Email.Value(user.GetEmail()),
		userMutation.Password.Value(enValue),
		userMutation.Salt.Value(pass.Salt()),
	}
	result, err := userMutation.WithContext(ctx).Where(wrapper...).UpdateSimple(mutations...)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, merr.ErrorUserNotFound("user not found")
	}

	userDo, err := userMutation.WithContext(ctx).Where(userMutation.ID.Eq(user.GetID()), userMutation.Email.Eq(user.GetEmail())).First()
	if err != nil {
		return nil, userNotFound(err)
	}

	if err = u.sendUserPassword(ctx, userDo, pass.PValue(), sendEmailFunc); err != nil {
		return nil, err
	}
	return userDo, nil
}

//go:embed template/welcome.html
var welcomeEmailBody string

func (u *userRepoImpl) sendUserPassword(ctx context.Context, user do.User, pass string, sendEmailFunc bo.SendEmailFun) error {
	if err := validate.CheckEmail(user.GetEmail()); err != nil {
		return nil
	}

	bodyParams := map[string]string{
		"Username":    user.GetEmail(),
		"Password":    pass,
		"RedirectURI": u.bc.GetAuth().GetOauth2().GetRedirectUri(),
	}
	emailBody, err := template.HtmlFormatter(welcomeEmailBody, bodyParams)
	if err != nil {
		return err
	}
	sendEmailParams := &bo.SendEmailParams{
		Email:       user.GetEmail(),
		Body:        emailBody,
		Subject:     "Welcome to the Moon Monitoring System.",
		ContentType: "text/html",
		RequestID:   uuid.New().String(),
	}
	// send email to user
	return sendEmailFunc(ctx, sendEmailParams)
}

// GetTeamsByUserID Gets all the teams to which the user belongs
func (u *userRepoImpl) GetTeamsByUserID(ctx context.Context, userID uint32) ([]do.Team, error) {
	mutation := getMainQuery(ctx, u)
	userQuery := mutation.User
	user, err := userQuery.WithContext(ctx).Where(userQuery.ID.Eq(userID)).Preload(userQuery.Teams.Where(mutation.Team.ID)).First()
	if err != nil {
		return nil, userNotFound(err)
	}
	if len(user.GetTeams()) == 0 {
		return nil, nil
	}
	teamIds := slices.Map(user.GetTeams(), func(team do.Team) uint32 { return team.GetID() })
	teamQuery := mutation.Team
	wrappers := []gen.Condition{
		teamQuery.ID.In(teamIds...),
	}

	teamDos, err := teamQuery.WithContext(ctx).Where(wrappers...).Preload(teamQuery.Admins, teamQuery.Leader).Order(teamQuery.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	teams := slices.Map(teamDos, func(team *system.Team) do.Team { return team })
	return teams, nil
}

func (u *userRepoImpl) UpdateUserInfo(ctx context.Context, user do.User) error {
	userMutation := getMainQuery(ctx, u).User
	_, err := userMutation.WithContext(ctx).
		Where(userMutation.ID.Eq(user.GetID())).
		UpdateSimple(
			userMutation.Nickname.Value(user.GetNickname()),
			userMutation.Avatar.Value(user.GetAvatar()),
			userMutation.Gender.Value(int8(user.GetGender())),
		)
	if err != nil {
		return err
	}

	return nil
}

//go:embed template/password_changed.html
var passwordChangedEmailBody string

// UpdatePassword updates the user's password in the database
func (u *userRepoImpl) UpdatePassword(ctx context.Context, updateUserPasswordInfo *bo.UpdateUserPasswordInfo) error {
	userDo, err := u.FindByID(ctx, updateUserPasswordInfo.UserID)
	if err != nil {
		return err
	}
	if err := validate.CheckEmail(userDo.GetEmail()); err != nil {
		return err
	}
	defer func() {
		if updateUserPasswordInfo.SendEmailFun == nil {
			return
		}
		bodyParams := map[string]string{
			"Email":       userDo.GetEmail(),
			"Password":    updateUserPasswordInfo.OriginPassword,
			"RedirectURI": u.bc.GetAuth().GetOauth2().GetRedirectUri(),
		}
		body, err := template.HtmlFormatter(passwordChangedEmailBody, bodyParams)
		if err != nil {
			u.helper.WithContext(ctx).Errorw("msg", "format email body error", "error", err)
			return
		}

		sendEmailParams := &bo.SendEmailParams{
			Email:       userDo.GetEmail(),
			Body:        body,
			Subject:     "Password reset.",
			ContentType: "text/html",
			RequestID:   uuid.New().String(),
		}
		if err := updateUserPasswordInfo.SendEmailFun(ctx, sendEmailParams); err != nil {
			u.helper.WithContext(ctx).Errorw("msg", "send email error", "error", err)
			return
		}
	}()
	userMutation := getMainQuery(ctx, u).User
	mutations := []field.AssignExpr{
		userMutation.Password.Value(updateUserPasswordInfo.Password),
		userMutation.Salt.Value(updateUserPasswordInfo.Salt),
	}

	// Update password and salt fields
	_, err = userMutation.WithContext(ctx).
		Where(userMutation.ID.Eq(updateUserPasswordInfo.UserID)).
		UpdateSimple(mutations...)

	return err
}

func (u *userRepoImpl) UpdateUserStatus(ctx context.Context, req *bo.UpdateUserStatusRequest) error {
	if len(req.UserIds) == 0 {
		return nil
	}
	userMutation := getMainQuery(ctx, u).User
	_, err := userMutation.WithContext(ctx).
		Where(userMutation.ID.In(req.UserIds...)).
		UpdateSimple(userMutation.Status.Value(req.Status.GetValue()))
	return err
}

func (u *userRepoImpl) UpdateUserPosition(ctx context.Context, req bo.UpdateUserPosition) error {
	userMutation := getMainQuery(ctx, u).User
	_, err := userMutation.WithContext(ctx).
		Where(userMutation.ID.Eq(req.GetUser().GetID())).
		UpdateSimple(userMutation.Position.Value(req.GetPosition().GetValue()))
	return err
}

func (u *userRepoImpl) List(ctx context.Context, req *bo.UserListRequest) (*bo.UserListReply, error) {
	query := getMainQuery(ctx, u)
	userQuery := query.User
	wrapper := userQuery.WithContext(ctx)

	if !validate.TextIsNull(req.Keyword) {
		ors := []gen.Condition{
			userQuery.Nickname.Like(req.Keyword),
			userQuery.Username.Like(req.Keyword),
			userQuery.Remark.Like(req.Keyword),
			userQuery.Email.Eq(req.Keyword),
			userQuery.Phone.Eq(req.Keyword),
		}
		wrapper = wrapper.Where(userQuery.Or(ors...))
	}
	if len(req.Status) > 0 {
		status := slices.Map(req.Status, func(statusItem vobj.UserStatus) int8 { return statusItem.GetValue() })
		wrapper = wrapper.Where(userQuery.Status.In(status...))
	}
	if len(req.Position) > 0 {
		position := slices.Map(req.Position, func(positionItem vobj.Role) int8 { return positionItem.GetValue() })
		wrapper = wrapper.Where(userQuery.Position.In(position...))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	users, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return req.ToListUserReply(users), nil
}
