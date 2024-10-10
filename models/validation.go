package models

type GroupInfo struct {
	Name  string `json:"name" binding:"required"`
	Realm string `json:"realm" binding:"required"`
}

type CreateUserInfo struct {
	Username  string `json:"username" binding:"required"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Realm     string `json:"realm" binding:"required"`
}

type UpdateUserInfo struct {
	Username        *string              `json:"username"`
	FirstName       *string              `json:"firstName"`
	LastName        *string              `json:"lastName,omitempty"`
	Email           *string              `json:"email"`
	PhoneNumber     *string              `json:"phoneNumber"`
	Attributes      *map[string][]string `json:"attributes,omitempty"`
	RequiredActions *[]string            `json:"requiredActions"`
	Enabled         *bool                `json:"enabled"`
}

type GetUserInfo struct {
	ID               *string              `json:"id,omitempty"`
	CreatedTimestamp *int64               `json:"createdTimestamp,omitempty"`
	Username         *string              `json:"username,omitempty"`
	Enabled          *bool                `json:"enabled"`
	FirstName        *string              `json:"firstName"`
	LastName         *string              `json:"lastName"`
	Email            *string              `json:"email"`
	Realm            *string              `json:"realm"`
	PhoneNumber      *string              `json:"phoneNumber"`
	Groups           *string              `json:"groups,omitempty"`
	Roles            *string              `json:"roles,omitempty"`
	Attributes       *map[string][]string `json:"attributes,omitempty"`
	Account          *string              `json:"Account,omitempty"`
	AccountId        *string              `json:"AccountId,omitempty"`
	OpenId           *string              `json:"OpenId,omitempty"`
	RequiredActions  *[]string            `json:"requiredActions,omitempty"`
	CreateDate       *string              `json:"createDate"`
	Creator          *string              `json:"creator"`
	ModifyDate       *string              `json:"modifyDate"`
	Modifier         *string              `json:"modifier"`
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
	TenantId    string    `json:"tenantId,omitempty"`
	DefaultRole bool      `json:"defaultRole"`
	Realm       string    `json:"realm,omitempty"`
	AuthId      *[]string `json:"authId,omitempty"`
	CreateDate  *string   `json:"createDate,omitempty"`
	Creator     *string   `json:"creator,omitempty"`
	ModifyDate  *string   `json:"modifyDate,omitempty"`
	Modifier    *string   `json:"modifier,omitempty"`
}

type AutuhorityInfo struct {
	ID         string  `json:"id" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	URL        *string `json:"url,omitempty"`
	Method     *string `json:"method,omitempty"`
	Use        *bool   `json:"useYn,omitempty"`
	Realm      string  `json:"realm,omitempty"`
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
	Realm        string  `json:"realm" binding:"required"`
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
	Realm string `json:"realm,omitempty"`
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

/* Swagger 출력용 구조체 */
type Id struct {
	Id string `json:"id" binding:"required"`
}

type Active struct {
	Active bool `json:"active" binding:"required"`
}

type CredentialRepresentation struct {
	Id             string `json:"id" binding:"required"`
	Type           string `json:"type" binding:"required"`
	CreatedDate    int    `json:"createdDate" binding:"required"`
	CredentialData string `json:"credentialData" binding:"required"`
	UserLabel      string `json:"userLabel"`
}

type GroupData struct {
	Id   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Path string `json:"path" binding:"required"`
}

type SesstionData struct {
	Id         string `json:"id" binding:"required"`
	Username   string `json:"username" binding:"required"`
	UserId     string `json:"userId" binding:"required"`
	IpAddress  string `json:"ipAddress" binding:"required"`
	Start      int    `json:"start" binding:"required"`
	LastAccess int    `json:"lastAccess" binding:"required"`
	Clients    []struct {
		ClientId_Text string `json:"ClientId_Text" binding:"required"`
	} `json:"clients" binding:"required"`
}

type UserIdProviderData struct {
	IdentityProvider string `json:"identityProvider" binding:"required"`
	UserId           string `json:"userId" binding:"required"`
	UserName         string `json:"userName" binding:"required"`
}

type SecretGroupInput struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	RoleId      *[]string `json:"roleId,omitempty"`
	UserId      *[]string `json:"userId,omitempty"`
}

type SecretGroupUpdate struct {
	Description string    `json:"description" binding:"required"`
	RoleId      *[]string `json:"roleId,omitempty"`
	UserId      *[]string `json:"userId,omitempty"`
}

type SecretData struct {
	Url        string `json:"url" binding:"required"`
	CreateDate string `json:"createDate" binding:"required"`
	Creator    string `json:"creator" binding:"required"`
	ModifyDate string `json:"modifyDate" binding:"required"`
	Modifier   string `json:"modifier" binding:"required"`
}

type SecretInput struct {
	Data struct {
		Foo  string `json:"foo" binding:"required"`
		Test string `json:"test" binding:"required"`
	} `json:"data" binding:"required"`
	URL string `json:"url" binding:"required"`
}

type SecretMetadata struct {
	Created_time    string `json:"created_time" binding:"required"`
	Current_version string `json:"current_version" binding:"required"`
	Max_versions    string `json:"max_versions" binding:"required"`
	Oldest_version  string `json:"oldest_version" binding:"required"`
	Updated_time    string `json:"updated_time" binding:"required"`
	Versions        struct {
		Version_number struct {
			Created_time  string `json:"created_time" binding:"required"`
			Deletion_time string `json:"deletion_time" binding:"required"`
			Destroyed     bool   `json:"destroyed" binding:"required"`
		} `json:"versversion_numberions" binding:"required"`
	} `json:"versions" binding:"required"`
}

type SecretVersion struct {
	Versions []string `json:"versions" binding:"required"`
}

type MetricCount struct {
	Users        int `json:"users" binding:"required"`
	Groups       int `json:"groups" binding:"required"`
	Applications int `json:"applications" binding:"required"`
	Roles        int `json:"roles" binding:"required"`
	Authorities  int `json:"authorities" binding:"required"`
}

type MetricAppItem struct {
	Client1 int    `json:"client1"`
	Client2 int    `json:"client2"`
	Date    string `json:"date"`
}

type MetricLogItem struct {
	ClientId  string `json:"clientId"`
	Username  string `json:"username"`
	EventDate string `json:"eventDate"`
}

type GetServiceAccount struct {
	ID         *string `json:"id,omitempty"`
	Username   *string `json:"username,omitempty"`
	ClientId   *string `json:"clientId,omitempty"`
	RealmId    *string `json:"realmId,omitempty"`
	Enabled    *bool   `json:"enabled"`
	Roles      *string `json:"roles,omitempty"`
	Secret     *string `json:"secret,omitempty"`
	Account    *string `json:"account,omitempty"`
	AccountId  *string `json:"accountId,omitempty"`
	CreateDate string  `json:"createDate"`
	Creator    string  `json:"creator"`
	ModifyDate string  `json:"modifyDate"`
	Modifier   string  `json:"modifier"`
}

type GetServiceAccountInfo struct {
	ID               *string `json:"id,omitempty"`
	CreatedTimestamp *int64  `json:"createdTimestamp,omitempty"`
	Username         *string `json:"username,omitempty"`
	Enabled          *bool   `json:"enabled"`
	Realm            *string `json:"realm"`
	Roles            *string `json:"roles,omitempty"`
	Account          *string `json:"Account,omitempty"`
	AccountId        *string `json:"AccountId,omitempty"`
	CreateDate       string  `json:"createDate"`
	Creator          string  `json:"creator"`
	ModifyDate       string  `json:"modifyDate"`
	Modifier         string  `json:"modifier"`
}

type ClientSecret struct {
	Type  *string `json:"type"`
	Value *string `json:"value"`
}

type CreateServiceAccount struct {
	ClientId string `json:"clientId"`
}

type UpdateServiceAccount struct {
	Enabled  bool   `json:"enabled"`
	ClientId string `json:"clientId"`
}
