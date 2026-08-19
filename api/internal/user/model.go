package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// [CHANGE] เปลี่ยน model ให้ตรงกับ table users จริง
type User struct {
	ID                uuid.UUID `gorm:"primaryKey;default:gen_random_uuid()"`
	Email             string
	Name              *string
	Bio               *string
	UID               string
	PreferredUsername *string
	LastSignedInAt    *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
}

// [CHANGE] เพิ่ม keycloak user จาก claim
type KeycloakUser struct {
	UID               string
	Email             string
	Name              string
	PreferredUsername string
}
