package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trip struct {
	ID          primitive.ObjectID `bson:"_id"`
	Trip_ID     *string            `json:"trip_id"`
	Name        *string            `json:"trip_name"`
	Description *string            `json:"description"`
	Members     *[]string          `json:"members"`
	IsDeleted   *bool              `bson:"is_deleted" json:"is_deleted"`
	Creator_ID  *string            `json:"creator_id"`
	Invite_Code *string            `json:"invite_code"`
	Created_At  time.Time          `json:"created_at"`
}
