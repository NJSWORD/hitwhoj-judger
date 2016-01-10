package runInstanceModel

import "gopkg.in/mgo.v2/bson"

type Run struct {
	Id     bson.ObjectId
	Rid    int
	Source string
	Lang   string
	Pid    int
}
