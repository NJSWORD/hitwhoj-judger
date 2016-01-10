package CorCppCompiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lackofdream/oj/runInstanceModel"
	"os"
	"os/exec"
	"time"
)

type CEMessage struct {
	Type    int
	Message string
}

func delay(run runInstanceModel.Run) {
	time.Sleep(time.Second * 5)
	infoOb := CEMessage{0, "timeout"}
	info, _ := json.Marshal(infoOb)
	ioutil.WriteFile(fmt.Sprintf("results/%d", run.Rid), info, 0644)
	os.Exit(1)
}

func Compile(run runInstanceModel.Run, srcFilePath, dstFilePath string) bool {
	var command *exec.Cmd
	if run.Lang == "c++" {
		command = exec.Command("g++", "-w", "-O2", "--static", "-lm", "-DONLINE_JUDGE", "-o", dstFilePath, srcFilePath)
	} else if run.Lang == "c" {
		command = exec.Command("gcc", "-w", "-O2", "--static", "-lm", "-DONLINE_JUDGE", "-o", dstFilePath, srcFilePath)
	} else if run.Lang == "java" {
		command = exec.Command("laaas")
	} else {
		command = exec.Command("laaas")
	}
	stderr, _ := command.StderrPipe()
	command.Start()
	go delay(run)
	defer command.Wait()
	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	s := buf.String()
	if len(s) != 0 {
		infoOb := CEMessage{0, s}
		info, _ := json.Marshal(infoOb)
		ioutil.WriteFile(fmt.Sprintf("results/%d", run.Rid), info, 0644)
		return false
	}
	return true
}
