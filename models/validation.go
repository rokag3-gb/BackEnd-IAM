package models

type GroupInfo struct {
	Name string `json:"name" binding:"required"`
}

type CreateUserInfo struct {
	Username  string `json:"username" binding:"required"`
	FirstName string `json:"firstName"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type UpdateUserInfo struct {
	Username        string   `json:"username"`
	FirstName       string   `json:"firstName"`
	Email           string   `json:"email"`
	RequiredActions []string `json:"requiredActions"`
	Enabled         bool     `json:"enabled"`
}

type GetUserInfo struct {
	ID               *string   `json:"id,omitempty"`
	CreatedTimestamp *int64    `json:"createdTimestamp,omitempty"`
	Username         *string   `json:"username,omitempty"`
	Enabled          *bool     `json:"enabled,omitempty"`
	EmailVerified    *bool     `json:"emailVerified,omitempty"`
	FirstName        *string   `json:"firstName,omitempty"`
	Email            *string   `json:"email,omitempty"`
	RequiredActions  *[]string `json:"requiredActions,omitempty"`
}

type ResetUserPasswordInfo struct {
	Password        string `json:"password" binding:"required,eqfield=PasswordConfirm"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
	Temporary       bool   `json:"temporary"`
}

type RolesInfo struct {
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Use  string `json:"useYn,omitempty"`
}

type AutuhorityInfo struct {
	ID     string `json:"id" binding:"required"`
	Name   string `json:"name" binding:"required"`
	URL    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
	Use    string `json:"useYn,omitempty"`
}

type GroupItem struct {
	ID           string `json:"id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	CountMembers int    `json:"countMembers" binding:"required"`
}

type SecretGroup struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type AutuhorityUse struct {
	Use string `json:"useYn" binding:"required"`
}
