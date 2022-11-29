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
	Groups           *string   `json:"groups,omitempty"`
	Roles            *string   `json:"roles,omitempty"`
	OpenId           *string   `json:"OpenId,omitempty"`
	RequiredActions  *[]string `json:"requiredActions,omitempty"`
	CreateDate       *string   `json:"createDate"`
	Creator          *string   `json:"creator"`
	ModifyDate       *string   `json:"modifyDate"`
	Modifier         *string   `json:"modifier"`
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
	ID          string    `json:"id" binding:"required"`
	Name        *string   `json:"name" binding:"required"`
	Use         bool      `json:"useYn,omitempty"`
	DefaultRole bool      `json:"defaultRole,omitempty"`
	AuthId      *[]string `json:"authId,omitempty"`
	CreateDate  *string   `json:"createDate"`
	Creator     *string   `json:"creator"`
	ModifyDate  *string   `json:"modifyDate"`
	Modifier    *string   `json:"modifier"`
}

type AutuhorityInfo struct {
	ID         string  `json:"id" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	URL        *string `json:"url,omitempty"`
	Method     *string `json:"method,omitempty"`
	Use        *bool   `json:"useYn,omitempty"`
	CreateDate *string `json:"createDate"`
	Creator    *string `json:"creator"`
	ModifyDate *string `json:"modifyDate"`
	Modifier   *string `json:"modifier"`
}

type MenuAutuhorityInfo struct {
	Name   string  `json:"name" binding:"required"`
	URL    *string `json:"url,required"`
	Method *string `json:"method,required"`
}

type GroupItem struct {
	ID           string  `json:"id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	CountMembers int     `json:"countMembers" binding:"required"`
	CreateDate   *string `json:"createDate"`
	Creator      *string `json:"creator"`
	ModifyDate   *string `json:"modifyDate"`
	Modifier     *string `json:"modifier"`
}

type SecretGroupItem struct {
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description" binding:"required"`
	RoleId      *[]string     `json:"roleId,omitempty"`
	UserId      *[]string     `json:"userId,omitempty"`
	CreateDate  *string       `json:"createDate"`
	Creator     *string       `json:"creator"`
	ModifyDate  *string       `json:"modifyDate"`
	Modifier    *string       `json:"modifier"`
	Secrets     *[]SecretItem `json:"secrets,omitempty"`
}

type SecretURL struct {
	URL *string `json:"url"`
}

type SecretGroupResponse struct {
	Description string   `json:"description" binding:"required"`
	Roles       []IdItem `json:"roles" binding:"required"`
	Users       []IdItem `json:"users" binding:"required"`
	CreateDate  *string  `json:"createDate"`
	Creator     *string  `json:"creator"`
	ModifyDate  *string  `json:"modifyDate"`
	Modifier    *string  `json:"modifier"`
}

type IdItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type MetricItem struct {
	Key   string `json:"key"`
	Value int    `json:"valse"`
}

type SecretItem struct {
	SecretGroup string  `json:"-"`
	Name        string  `json:"name" binding:"required"`
	Url         *string `json:"url" binding:"required"`
	CreateDate  *string `json:"createDate"`
	Creator     *string `json:"creator"`
	ModifyDate  *string `json:"modifyDate"`
	Modifier    *string `json:"modifier"`
}

type AutuhorityUse struct {
	Use bool `json:"useYn" binding:"required"`
}
