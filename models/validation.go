package models

type GroupInfo struct {
	Name string `json:"name" binding:"required"`
}

type CreateUserInfo struct {
	Username  string `json:"username" binding:"required"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type UpdateUserInfo struct {
	Username        string   `json:"username" binding:"required"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	Email           string   `json:"email" binding:"required"`
	RequiredActions []string `json:"requiredActions"`
}

type GetUserInfo struct {
	ID               *string   `json:"id,omitempty"`
	CreatedTimestamp *int64    `json:"createdTimestamp,omitempty"`
	Username         *string   `json:"username,omitempty"`
	Enabled          *bool     `json:"enabled,omitempty"`
	EmailVerified    *bool     `json:"emailVerified,omitempty"`
	FirstName        *string   `json:"firstName,omitempty"`
	LastName         *string   `json:"lastName,omitempty"`
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
}

type AutuhorityInfo struct {
	ID     string `json:"id" binding:"required"`
	Name   string `json:"name" binding:"required"`
	URL    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
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
