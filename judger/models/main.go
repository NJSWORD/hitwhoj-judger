package models

import "gopkg.in/mgo.v2/bson"


type Problem struct {
	Id        bson.ObjectId `bson:"_id"`
	Pid       int           `bson:"pid"`
	Statement string        `bson:"statement"`
	In        bson.ObjectId `bson:"in"`
	Out       bson.ObjectId `bson:"out"`
	Is_spj    bool          `bson:"is_spj"`
	Spj       bson.ObjectId `bson:"spj"`
}

type Run struct {
	Id       bson.ObjectId `bson:"_id"`
	Rid      int           `bson:"rid"`
	Source   string        `bson:"source"`
	Lang     string        `bson:"lang"`
	Pid      int           `bson:"pid"`
	Status   int           `bson:"status"`
	Data     string        `bson:"data"` // CE信息等
	Time     int           `bson:"time"`
	Memory   int           `bson:"memory"`
	Date     int           `bson:"date"`
	Username int           `bson:"username"`
}
