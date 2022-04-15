package models

type CreateGroupInfo struct {
	Name string `json:"name" binding:"required"`
}
