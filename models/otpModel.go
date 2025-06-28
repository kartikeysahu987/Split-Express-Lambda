package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OTP struct {
	ID        primitive.ObjectID `bson:"_id"`
	Email     string             `json:"email" bson:"email"`
	OTP       string             `json:"otp" bson:"otp"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	Used      bool               `json:"used" bson:"used"`
}

type OTPRequest struct {
	Email string `json:"email" validate:"email,required"`
}

type OTPVerification struct {
	Email string `json:"email" validate:"email,required"`
	OTP   string `json:"otp" validate:"required,len=6"`
}
