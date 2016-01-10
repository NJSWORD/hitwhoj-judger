package problemModel

import "gopkg.in/mgo.v2/bson"

type Problem struct {
	Id bson.ObjectId
	Pid int
	Statement string
	In bson.ObjectId
	Out bson.ObjectId
	Is_spj bool
	Spj bson.ObjectId
}