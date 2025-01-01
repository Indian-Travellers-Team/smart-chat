package models

import (
	"gorm.io/gorm"
)

type Button struct {
	gorm.Model
	Name string `gorm:"type:varchar(100); not null"`
	Text string `gorm:"type:varchar(255); not null"`
}
