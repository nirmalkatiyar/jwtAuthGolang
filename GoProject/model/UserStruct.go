package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id           primitive.ObjectID `bson:"id"`
	FirstName    *string            `json:"first_name" validate:"required,min=2,max=70"`
	LastName     *string            `json:"last_name"`
	Password     *string            `json:"password" validate:"required,min=8"`
	Email        *string            `json:"email" validate:"email,required"`
	Phone        *string            `json:"phone" validate:"required,min=10"`
	Token        *string            `json:"token"`
	RefreshToken *string            `json:"refresh_token"`
	UserType     *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	UserId       string             `json:"user_id"`
}
