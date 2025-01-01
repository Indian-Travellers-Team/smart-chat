package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name           string    `gorm:"type:varchar(100);not null"`
	Mobile         string    `gorm:"type:varchar(15); not null"`
	OTP            string    `gorm:"type:varchar(4); not null"`
	AccessToken    string    `gorm:"type:varchar(255); not null"`
	AccessExpireAt time.Time `gorm:"not null"`
	Sessions       []Session
}

type Session struct {
	gorm.Model
	UserID        uint   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User          User   `gorm:"foreignKey:UserID;references:ID"`
	AuthToken     string `gorm:"type:varchar(255)"`
	ExpireAt      time.Time
	Conversations []Conversation `gorm:"foreignKey:SessionID;"`
}
