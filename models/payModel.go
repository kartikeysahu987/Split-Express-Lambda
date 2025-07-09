package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct{
	ID 				primitive.ObjectID		`bson:"_id"`
	Trip_ID			*string					`json:"trip_id"`
	PayerName		*string					`json:"payer_name"`
	ReciverName		*string					`json:"reciever_name"`
	Amount			*string					`json:"amount"`
	Description		*string					`json:"description"`
	IsDeleted		*bool					`json:"is_deleted"`
	Type			*string					`json:"type"`
	Created_At		time.Time				`json:"created_at"`
}