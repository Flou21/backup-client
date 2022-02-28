package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	TARGET_MONGO     = "mongo"
	TARGET_MYSQL     = "mysql"
	TARGET_CASSANDRA = "cassandra"
)

type Target struct {
	ID                     primitive.ObjectID `bson:"_id" json:"id"`
	Type                   string             `bson:"type" json:"type"`
	Ip                     string             `bson:"ip" json:"ip"`
	Port                   int64              `bson:"port" json:"port"`
	Name                   string             `bson:"name" json:"name"`
	Interval               int64              `json:"interval" bson:"interval"`
	Username               string             `bson:"username" json:"username"`
	Password               string             `bson:"password" json:"password"`
	Database               string             `bson:"database" json:"database"`
	AuthenticationDatabase string             `bson:"authentication_database,omitempty" json:"authenticationDatabase"`
}
