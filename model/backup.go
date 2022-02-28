package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Backup struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Target    *Target            `json:"target" bson:"target"`
	Path      string             `json:"path" bson:"path"`
	Size      int64              `json:"size" bson:"size"`
}
