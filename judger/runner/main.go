package runner

import (
	"bytes"
	"os"
	"os/exec"
	"log"

	"lackofdream/oj/judger/models"
	"lackofdream/oj/judger/languages"
	"time"
)

const (
	WAIT_FOR_JUDGE = 0
	COMPILE_ERROR = 1
	WAIT_FOR_RUN = 2
	RUN_TIME_ERROR = 3
	TIME_LIMIT_EXCEED = 4
	WAIT_FOR_VALIDATE = 5
	SOURCE_NOT_FOUND = 6
)

func compile(cmd *exec.Cmd, c chan <- int) {
	cmd.Run()
	c <- 1
}

func runCompile(state *models.Run, cmd ...string) {
	command := exec.Command(cmd[0], cmd[1:]...)
	stdErr, _ := command.StderrPipe()
	finC := make(chan int)
	go compile(command, finC)
	fin := 0
	for {
		if fin == 0 {
			select {
			case <-time.After(5 * time.Second):
				log.Println("Compile timeout, killing it...")
				state.Status = COMPILE_ERROR
				state.Data = "timeout\n"
				command.Process.Kill()
				log.Println("compiler killed")
				return
			case <-finC:
				fin = 1
			}
		} else {
			break
		}
	}
	close(finC)
	stdErrBuf := new(bytes.Buffer)
	stdErrBuf.ReadFrom(stdErr)
	stdErrStr := stdErrBuf.String()
	if len(stdErrStr) != 0 {
		state.Status = COMPILE_ERROR
		state.Data = stdErrStr
	} else {
		state.Status = WAIT_FOR_RUN
	}
}

func Compile(state *models.Run) {

	log.Println("Starting compiling...")
	if lang, ok := languages.Languages[state.Lang]; ok {
		if _, err := os.Stat(lang.SourceFile); err == nil {
			runCompile(state, lang.CompileCmd...)
		} else {
			state.Status = SOURCE_NOT_FOUND
		}
	} else {
		state.Status = COMPILE_ERROR
		state.Data = "unknown language\n"
	}
	log.Printf("Compiling finished, status: %d\n", state.Status)
	return
}
