package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"lackofdream/oj/runInstanceModel"
	"lackofdream/oj/problemModel"
	"io/ioutil"
	"strings"
	"bytes"
)


func handleFatalErr(err error, position string) {
	if err != nil {
		log.Fatal("error in", position)
		log.Println(err.Error())
		os.Exit(1)
	}
}

func handleNormalErr(err error, position string) bool {
	if err != nil {
		log.Printf("error in %s\n", position)
		log.Println(err.Error())
		return true
	}
	return false
}

func consume(runID int) {

	// 连接 Mongodb
	session, err := mgo.Dial(os.Args[4])
	if handleNormalErr(err, "connect to mongodb") {
		return
	}

	//  获取 runInstance
	db := session.DB("oj")
	runsCollection := db.C("runs")
	problemsCollection := db.C("problems")
	gridFS := db.GridFS("fs")
	var runInstance runInstanceModel.Run
	err = runsCollection.Find(bson.M{"rid": runID}).One(&runInstance)
	if handleNormalErr(err, "find runInstance in mongodb") {
		return
	}

	// 获取问题数据
	problemID := runInstance.Pid
	var problem problemModel.Problem
	err = problemsCollection.Find(bson.M{"pid": problemID}).One(&problem)
	if handleNormalErr(err, "fetch problem from mongodb") {
		return
	}

	// 创建 /tmp/oj_run/runID 工作目录，用于docker挂载
	workDir := fmt.Sprintf("/tmp/oj_run/%d", runID)
	cmd := exec.Command("mkdir", "-p", workDir, workDir + "/public", workDir + "/private")
	err = cmd.Run()
	if handleNormalErr(err, "create run folder") {
		return
	}

	// 获取输入、输出文件
	inDataID := problem.In
	inF, err := gridFS.OpenId(inDataID)
	if handleNormalErr(err, "fetch in.txt from GridFS") {
		return
	}
	inBuf := new(bytes.Buffer)
	inBuf.ReadFrom(inF)
	ioutil.WriteFile(workDir + "/public/in.txt", inBuf.Bytes(), 0644)

	outDataID := problem.Out
	outF, err := gridFS.OpenId(outDataID)
	if handleNormalErr(err, "fetch out.txt from GridFS") {
		return
	}
	outBuf := new(bytes.Buffer)
	outBuf.ReadFrom(outF)
	ioutil.WriteFile(workDir + "/private/out.txt", outBuf.Bytes(), 0644)

	// Special Judge Related
	if problem.Is_spj {
		spjID := problem.Spj
		spjF, err := gridFS.OpenId(spjID)
		if handleNormalErr(err, "fetch spj from GridFS") {
			return
		}
		spjBuf := new(bytes.Buffer)
		spjBuf.ReadFrom(spjF)
		ioutil.WriteFile(workDir + "/public/spj", spjBuf.Bytes(), 0744)
	}

	// 创建源代码文件
	var fileName string
	switch strings.ToLower(runInstance.Lang) {
	case "c":
		fileName = "Main.c"
	case "c++":
		fileName = "Main.cpp"
	case "java":
		fileName = "Main.java"
	}
	ioutil.WriteFile(workDir + "/public/" + fileName, []byte(runInstance.Source), 0644)

	// TODO: docker run ...
	log.Println("next step: create a docker container")
}

func main() {
	// 检测参数个数，参数个数小于4则显示 usage 信息
	if len(os.Args) < 5 {
		fmt.Printf("usage: %s REDIS_HOST REDIS_PORT REDIS_KEY MONGO_URL\n", os.Args[0])
		return
	}
	// 连接 Redis 消息队列
	conn, err := redis.Dial("tcp", os.Args[1] + ":" + os.Args[2])
	handleFatalErr(err, "connect to redis")
	defer conn.Close()
	// 接受消息
	for {
		run, err := conn.Do("BLPOP", os.Args[3], "0")
		handleFatalErr(err, "fetch message from redis")
		// 获取 int 类型的 runID
		runID, err := redis.Int(run.([]interface{})[1], err)
		handleFatalErr(err, "convert to runID from redis data")
		log.Printf("get runID: %v\n", runID)
		consume(runID)
	}
}
