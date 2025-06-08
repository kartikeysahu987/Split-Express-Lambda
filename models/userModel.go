package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_Name    *string            `json:"first_name"`
	Last_Name     *string            `json:"last_name"`
	Password      *string            `json:"password"`
	Email         *string            `json:"email" validate:"email,required"`
	Phone         *string            `json:"phone" validate:"required"`
	Token         *string            `json:"token"`
	User_type     *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	User_id       *string            `json:"user_id"`
}

// GetID implements mgm.Model.
func (u *User) GetID() interface{} {
	panic("unimplemented")
}

// PrepareID implements mgm.Model.
func (u *User) PrepareID(id interface{}) (interface{}, error) {
	panic("unimplemented")
}

// SetID implements mgm.Model.
func (u *User) SetID(id interface{}) {
	panic("unimplemented")
}
