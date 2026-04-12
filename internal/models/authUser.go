package models

type AuthUser struct {
	UserID        uint    `gorm:"column:user_id;primaryKey;autoIncrement"`
	ZitadelUserID string  `gorm:"column:zitadel_user_id;type:varchar(255);unique;not null;index"`
	Name          *string `gorm:"column:name;type:varchar(255)"`
	Email         *string `gorm:"column:email;type:varchar(255);index"`
	RoleID        uint    `gorm:"column:role_id;not null"`
}

func (AuthUser) TableName() string {
	return "auth_users"
}
