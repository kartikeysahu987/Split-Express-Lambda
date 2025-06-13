package models

import "go.mongodb.org/mongo-driver/bson/primitive"


type Member struct {
	ID 				primitive.ObjectID		`bson:"_id"`
	Trip_ID			*string					`json:"trip_id"`
	Name			*string					`json:"name"`
	Uid				*string					`json:"uid"`
}