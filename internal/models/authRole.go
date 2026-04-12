package models

type AuthRole struct {
	RoleID uint   `gorm:"column:role_id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name;type:varchar(50);unique;not null"`
}

func (AuthRole) TableName() string {
	return "auth_roles"
}
