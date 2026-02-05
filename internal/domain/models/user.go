package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID                primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Email             string               `bson:"email" json:"email" validate:"required,email"`
	PasswordHash      string               `bson:"password_hash" json:"-"` // Never expose in JSON
	FirstName         string               `bson:"first_name" json:"first_name" validate:"required,min=2,max=50"`
	LastName          string               `bson:"last_name" json:"last_name" validate:"required,min=2,max=50"`
	Phone             string               `bson:"phone,omitempty" json:"phone,omitempty" validate:"omitempty,e164"`
	EmailVerified     bool                 `bson:"email_verified" json:"email_verified"`
	EmailVerifiedAt   *time.Time           `bson:"email_verified_at,omitempty" json:"email_verified_at,omitempty"`
	ProfileImageURL   string               `bson:"profile_image_url,omitempty" json:"profile_image_url,omitempty" validate:"omitempty,url"`
	WeddingIDs        []primitive.ObjectID `bson:"wedding_ids" json:"wedding_ids"` // References to weddings
	CreatedAt         time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time            `bson:"updated_at" json:"updated_at"`
	LastLoginAt       *time.Time           `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
	Status            string               `bson:"status" json:"status" validate:"oneof=active inactive suspended"`
	PreferredLanguage string               `bson:"preferred_language,omitempty" json:"preferred_language,omitempty"`
	Timezone          string               `bson:"timezone,omitempty" json:"timezone,omitempty"`
}

// UserStatus represents possible user statuses
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)
