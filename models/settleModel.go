package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Settle struct{
	ID 				primitive.ObjectID		`bson:"_id"`
	Trip_ID			*string					`json:"trip_id"`
	PayerName		*string					`json:"payer_name"`
	ReciverName		*string					`json:"reciever_name"`
	Amount			*string					`json:"amount"`
	// Description		*string					`json:"description"`
	Created_At		time.Time				`json:"created_at"`
}