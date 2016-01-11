package runner

import (
	"bytes"
	"os"
	"os/exec"
	"log"

	"lackofdream/oj/judger/models"
	"lackofdream/oj/judger/languages"
	"time"
	"syscall"
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
	err := cmd.Wait()
	if err == nil {
		c <- 1
	}
}

// cmd.Process.Kill() doesn't kill child processes.
// see in http://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
func runCompile(state *models.Run, cmd ...string) {
	command := exec.Command(cmd[0], cmd[1:]...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdErr, _ := command.StderrPipe()
	finC := make(chan int)
	defer close(finC)
	command.Start()
	pgid, _ := syscall.Getpgid(command.Process.Pid)
	go compile(command, finC)
	select {
	case <-time.After(5 * time.Second):
		log.Println("Compile timeout, kill it")
		state.Status = COMPILE_ERROR
		state.Data = "timeout\n"
		syscall.Kill(-pgid, 9)
		return
	case <-finC:
	}

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
