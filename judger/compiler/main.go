package CorCppCompiler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

type CEMessage struct {
	Type    int
	Message string
}

func delay() {
	time.Sleep(time.Second * 5)
	//	infoOb := CEMessage{0, "timeout"}
	//	info, _ := json.Marshal(infoOb)
	//	ioutil.WriteFile(fmt.Sprintf("results"), info, 0644)
	os.Exit(1)
}

func Compile(srcFilePath, dstFilePath string) bool {
	var command *exec.Cmd
	stderr, _ := command.StderrPipe()
	command.Start()
	go delay()
	defer command.Wait()
	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	s := buf.String()
	if len(s) != 0 {
		infoOb := CEMessage{0, s}
		info, _ := json.Marshal(infoOb)
		ioutil.WriteFile("filename", info, 0644)
		return false
	}
	return true
}
