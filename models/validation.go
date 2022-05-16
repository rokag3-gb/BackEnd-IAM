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
	Username        string   `json:"username"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName,omitempty"`
	Email           string   `json:"email"`
	RequiredActions []string `json:"requiredActions"`
	Enabled         bool     `json:"enabled"`
}

type GetUserInfo struct {
	ID               *string   `json:"id,omitempty"`
	CreatedTimestamp *int64    `json:"createdTimestamp,omitempty"`
	Username         *string   `json:"username,omitempty"`
	Enabled          *bool     `json:"enabled"`
	FirstName        *string   `json:"firstName"`
	LastName         *string   `json:"lastName"`
	Email            *string   `json:"email"`
	RequiredActions  *[]string `json:"requiredActions,omitempty"`
	CreateDate       *string   `json:"createDate"`
	CreateId         *string   `json:"createId"`
	ModifyDate       *string   `json:"modifyDate"`
	ModifyId         *string   `json:"modifyId"`
}

type ResetUserPasswordInfo struct {
	Password        string `json:"password" binding:"required,eqfield=PasswordConfirm"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
	Temporary       bool   `json:"temporary"`
}

type UserRolesInfo struct {
	ID        string `json:"id" binding:"required"`
	Username  string `json:"username" binding:"required"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required"`
	RoleList  string `json:"roleList" binding:"required"`
}

type RolesInfo struct {
	ID         string  `json:"id" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	Use        string  `json:"useYn,omitempty"`
	CreateDate *string `json:"createDate"`
	CreateId   *string `json:"createId"`
	ModifyDate *string `json:"modifyDate"`
	ModifyId   *string `json:"modifyId"`
}

type AutuhorityInfo struct {
	ID         string  `json:"id" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	URL        string  `json:"url,omitempty"`
	Method     string  `json:"method,omitempty"`
	Use        string  `json:"useYn,omitempty"`
	CreateDate *string `json:"createDate"`
	CreateId   *string `json:"createId"`
	ModifyDate *string `json:"modifyDate"`
	ModifyId   *string `json:"modifyId"`
}

type GroupItem struct {
	ID           string  `json:"id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	CountMembers int     `json:"countMembers" binding:"required"`
	CreateDate   *string `json:"createDate"`
	CreateId     *string `json:"createId"`
	ModifyDate   *string `json:"modifyDate"`
	ModifyId     *string `json:"modifyId"`
}

type SecretGroup struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type AutuhorityUse struct {
	Use string `json:"useYn" binding:"required"`
}
