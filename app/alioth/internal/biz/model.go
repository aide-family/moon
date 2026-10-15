package biz

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid")
)

type Account struct {
	UID          string
	Name         string
	Nickname     string
	Email        string
	Avatar       string
	Remark       string
	Status       string
	PasswordHash string
}

type Namespace struct {
	UID      string
	Name     string
	OwnerUID string
	Logo     string
	Remark   string
	Status   string
}

type Member struct {
	UID          string
	NamespaceUID string
	UserUID      string
	Email        string
	Name         string
	Status       string
	RoleUIDs     []string
	RoleNames    []string
}

type Role struct {
	UID          string
	NamespaceUID string
	Name         string
	Status       string
	IsSystem     bool
	Permissions  []string
}

type Permission struct {
	UID      string
	Name     string
	Resource string
	Action   string
}

type SelectItem struct {
	UID  string
	Name string
}

type Session struct {
	AccessToken  string
	RefreshToken string
	User         Account
}
