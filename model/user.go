package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	UserId       int `gorm:"type:smallint"`
	Online       bool
	LastSeenTime time.Time
}
