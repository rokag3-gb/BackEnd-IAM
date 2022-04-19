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
