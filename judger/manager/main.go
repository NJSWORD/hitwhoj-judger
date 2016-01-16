package main

import (
	"fmt"
	"log"
	"os"
	"flag"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"lackofdream/oj/judger/models"
	"lackofdream/oj/judger/languages"
	"io/ioutil"
	"strings"
	"bytes"
	"lackofdream/oj/judger/runner"
)


var (
	redisHost string
	redisPort string
	redisAddress string
	redisKey string
	mongoHost string
	mongoPort string
	mongoAddress string
	maxWorkers int
	judgeUID int
	judgeGID int

	redisConnection redis.Conn
	mongoSession *mgo.Session
)

func init() {
	flag.StringVar(&redisHost, "redis-host", "localhost", "IP address of Redis")
	flag.StringVar(&redisPort, "redis-port", "6379", "Port number of Redis")
	flag.StringVar(&redisKey, "redis-key", "runs", "Key in Redis which store run instance IDs")
	flag.StringVar(&mongoHost, "mongo-host", "localhost", "IP address of MongoDB")
	flag.StringVar(&mongoPort, "mongo-port", "27017", "Port number of MongoDB")
	flag.IntVar(&maxWorkers, "max-workers", 4, "Maximum number of workers")
	flag.IntVar(&judgeUID, "uid", 1000, "UID for judge user")
	flag.IntVar(&judgeGID, "gid", 1000, "GID for judge group")

	flag.Parse()
	redisAddress = redisHost + ":" + redisPort
	mongoAddress = mongoHost + ":" + mongoPort

	var err error

	log.Println("Starting judge client...")

	// 连接 Redis 消息队列
	log.Printf("Connecting to Redis: %s...\n", redisAddress)
	redisConnection, err = redis.Dial("tcp", redisAddress)
	handleFatalErr(err, "connect to Redis")
	log.Println("Connected to Redis")
	// 连接 Mongodb
	log.Printf("Connecting to MongoDB: %s...\n", mongoAddress)
	mongoSession, err = mgo.Dial(mongoAddress)
	handleFatalErr(err, "connect to MongoDB")
	log.Println("Connected to MongoDB")
}

func handleFatalErr(err error, position string) {
	if err != nil {
		log.Fatal("error in ", position)
		log.Println(err.Error())
		os.Exit(1)
	}
}

func handleNormalErr(err error, position string, successPrompt string) bool {
	if err != nil {
		log.Printf("error in %s\n", position)
		return true
	}
	log.Println(successPrompt)
	return false
}

func worker(wid int, jobs <-chan int) {
	for j := range jobs {
		log.Printf("Worker%d get runID %d\n", wid, j)
		err := work(j)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func getRunInstance(runID int) (models.Run, error) {
	runsCollection := mongoSession.DB("oj").C("runs")
	var runInstance models.Run
	err := runsCollection.Find(bson.M{"rid": runID}).One(&runInstance)
	return runInstance, err
}

func getProblem(pID int) (models.Problem, error) {
	problemsCollection := mongoSession.DB("oj").C("problems")
	var problem models.Problem
	err := problemsCollection.Find(bson.M{"pid": pID}).One(&problem)
	return problem, err
}

func createFileFromGridFS(path string, objectID bson.ObjectId, perm os.FileMode) error {
	gridFS := mongoSession.DB("oj").GridFS("fs")
	grid, err := gridFS.OpenId(objectID)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(grid)
	ioutil.WriteFile(path, buf.Bytes(), perm)
	return nil
}

func work(runID int) error {

	//  获取 runInstance
	runInstance, err := getRunInstance(runID)
	if handleNormalErr(err, "find runInstance in mongodb", "runInstance found") {
		return err
	}
	defer func() {
		log.Println("Updating runInstance in MongoDB...")
		err := mongoSession.DB("oj").C("runs").Update(bson.M{"_id": runInstance.Id}, bson.M{"$set": runInstance})
		if err != nil {
			log.Println("Failed to update runInstance")
			return
		} else {
			log.Println("runInstance Updated")
		}
	}()
	// 获取问题数据
	problem, err := getProblem(runInstance.Pid)
	if handleNormalErr(err, "fetch problem from mongodb", "problem found") {
		return err
	}

	// 创建 /tmp/oj_run/runID 工作目录
	workDir := fmt.Sprintf("/tmp/oj_run/%d", runID)
	err = os.MkdirAll(workDir, 0755)
	defer os.RemoveAll(workDir)
	if handleNormalErr(err, "create run folder", "working directory created") {
		return err
	}

	// 进入工作目录
	os.Chdir(workDir)

	// 获取输入、输出文件
	err = createFileFromGridFS("in.txt", problem.In, 0644)
	if handleNormalErr(err, "fetch in.txt from GridFS", "in.txt found") {
		return err
	}

	err = createFileFromGridFS("out.txt", problem.Out, 0644)
	if handleNormalErr(err, "fetch out.txt from GridFS", "out.txt found") {
		return err
	}
	// Special Judge Related
	if problem.Is_spj {
		err = createFileFromGridFS("spj", problem.Spj, 0744)
		if handleNormalErr(err, "fetch spj from GridFS", "special judge binary found") {
			return err
		}
	}

	// 创建源代码文件
	var fileName string
	lang := strings.ToLower(runInstance.Lang)
	fileName = languages.Languages[lang].SourceFile
	ioutil.WriteFile(fileName, []byte(runInstance.Source), 0644)
	log.Printf("%s created\n", fileName)

	log.Println("Now wait for compile")

	if err := runner.Compile(&runInstance); err != nil {
		return err
	}

	err = runner.Execute(&runInstance, problem.Time_limit, problem.Memory_limit, 1000, 1000)
	if err != nil {
		return err
	}
	return runner.Validate(&runInstance)
}

func main() {

	defer redisConnection.Close()
	defer mongoSession.Close()

	// spawn workers
	log.Println("Spawning workers...")
	jobs := make(chan int)
	for i := 0; i < maxWorkers; i++ {
		log.Printf("Spawning worker%d...\n", i)
		go worker(i, jobs)
	}
	log.Printf("Workers spawned, totally %d\n", maxWorkers)

	// 接受消息
	for {
		log.Println("Waiting for run instance...")
		run, err := redisConnection.Do("BLPOP", redisKey, "0")
		handleFatalErr(err, "fetch message from redis")
		// 获取 int 类型的 runID
		runID, err := redis.Int(run.([]interface{})[1], err)
		handleFatalErr(err, "convert to runID from redis data")
		log.Printf("get runID: %v\n", runID)
		jobs <- runID
	}
}
