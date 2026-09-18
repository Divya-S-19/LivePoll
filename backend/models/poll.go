package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Poll struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Question  string        `bson:"question" json:"question"`
	Options   []string      `bson:"options" json:"options"`
	CreatedBy string        `bson:"created_by" json:"created_by"`
	CreatedAt string        `bson:"created_at" json:"created_at"`
}

type Vote struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    bson.ObjectID `bson:"poll_id" json:"poll_id"`
	Option    string        `bson:"option" json:"option"`
	CreatedAt string        `bson:"created_at" json:"created_at"`
}