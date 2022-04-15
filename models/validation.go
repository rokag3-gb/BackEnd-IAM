package models

type GroupInfo struct {
	Name string `json:"name" binding:"required"`
}
