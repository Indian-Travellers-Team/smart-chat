package specification

import "gorm.io/gorm"

type Specification interface {
	Apply(*gorm.DB) *gorm.DB
}
