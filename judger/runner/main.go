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
	"io"
	"fmt"
	"strings"
	"io/ioutil"
	"errors"
	"crypto/md5"
)

const (
	WAIT_FOR_JUDGE = 0
	COMPILE_ERROR = 1
	RUN_TIME_ERROR = 2
	TIME_LIMIT_EXCEED = 3
	MEMORY_LIMIT_EXCEED = 4
	SOURCE_NOT_FOUND = 5
	ACCEPT = 6
	WRONG_ANSWER = 7
	PRESENTATION_ERROR = 8
)

func compile(cmd *exec.Cmd, c chan <- int) {
	err := cmd.Wait()
	if err == nil || !strings.Contains(err.Error(), "kill") {
		c <- 1
	}
}

func getStrFromPipe(pipe io.Reader, str *string) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(pipe)
	*str = buf.String()
}

// cmd.Process.Kill() doesn't kill child processes.
// see in http://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
func runCompile(state *models.Run, cmd string) error {
	command := exec.Command("bash", "-c", cmd)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdErr, err := command.StderrPipe()
	if err != nil {
		return err
	}
	var stdErrStr string
	finC := make(chan int)
	defer close(finC)
	command.Start()
	go getStrFromPipe(stdErr, &stdErrStr)
	pgid, err := syscall.Getpgid(command.Process.Pid)
	if err != nil {
		return err
	}
	go compile(command, finC)
	select {
	case <-time.After(5 * time.Second):
		log.Println("Compile timeout, kill it")
		state.Status = COMPILE_ERROR
		state.Data = "timeout\n"
		syscall.Kill(-pgid, 9)
		return errors.New("Compile timeout")
	case <-finC:
	}

	if len(stdErrStr) != 0 {
		state.Status = COMPILE_ERROR
		state.Data = stdErrStr
		return errors.New("Compile Error")
	}
	return nil
}

func Compile(state *models.Run) error {

	log.Println("Starting compiling...")
	if lang, ok := languages.Languages[state.Lang]; ok {
		if _, err := os.Stat(lang.SourceFile); err == nil {
			return runCompile(state, lang.CompileCmd)
		} else {
			state.Status = SOURCE_NOT_FOUND
			return errors.New("Compile Error")
		}
	} else {
		state.Status = COMPILE_ERROR
		state.Data = "unknown language\n"
		return errors.New("Unknown Language")
	}
}

func Execute(state *models.Run, timeLimit, memoryLimit, uid, gid int) error {

	log.Println("Executing Main...")

	cmdStr :=
	" lrun " +
	" --max-cpu-time " + fmt.Sprintf("%.6f", float32(timeLimit) * languages.Languages[state.Lang].TimeOffset / 1000) +
	" --max-real-time " + fmt.Sprintf("%.6f", float32(timeLimit) * languages.Languages[state.Lang].TimeOffset / 500) +
	" --max-memory " + fmt.Sprintf("%dk", int(float32(memoryLimit) * languages.Languages[state.Lang].MemoryOffset)) +
	" --syscalls '" + languages.Languages[state.Lang].Syscalls + "'" +
	" --remount-dev true " +
	" --network false " +
	" --uid " + fmt.Sprintf("%d", uid) +
	" --gid " + fmt.Sprintf("%d", gid) +
	" --reset-env true " + languages.Languages[state.Lang].RunCmd + " 0<in.txt " + " 1>user_out.txt " + " 2>user_err.txt " + " 3>lrun.txt "
	log.Printf("Command to run: %v", cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)
	return cmd.Run()
}

type lrunInfo struct {
	Memory   int // KB
	Time     int // ms
	Signaled int
	ExitCode int
	TermSig  int
	Exceed   string
}

func parseLRUN(path string) (lrunInfo, error) {
	info := lrunInfo{}
	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return info, err
	}
	rawText := string(rawBytes)
	rawTexts := strings.Split(rawText, "\n")
	var tmp string
	var tmpFloat float32
	fmt.Sscanf(rawTexts[0], "%s%d", &tmp, &(info.Memory))
	info.Memory = info.Memory / 1024
	fmt.Sscanf(rawTexts[1], "%s%f", &tmp, &tmpFloat)
	info.Time = int(tmpFloat * 1000)
	fmt.Sscanf(rawTexts[3], "%s%d", &tmp, &(info.Signaled))
	fmt.Sscanf(rawTexts[4], "%s%d", &tmp, &(info.ExitCode))
	fmt.Sscanf(rawTexts[5], "%s%d", &tmp, &(info.TermSig))
	fmt.Sscanf(rawTexts[6], "%s%s", &tmp, &(info.Exceed))

	return info, nil
}

func validateLRUN(state *models.Run, info lrunInfo) bool {
	switch info.Exceed {
	case "MEMORY":
		state.Status = MEMORY_LIMIT_EXCEED
		return false
	case "CPU_TIME":
		state.Status = TIME_LIMIT_EXCEED
		return false
	case "REAL_TIME":
		state.Status = TIME_LIMIT_EXCEED
		return false
	}
	if info.ExitCode != 0 || info.TermSig != 0 || info.Signaled != 0 {
		state.Status = RUN_TIME_ERROR
		return false
	}
	return true
}

func diff(path1, path2 string) (bool, error) {
	file1, err := ioutil.ReadFile(path1)
	if err != nil {
		return false, err
	}
	file2, err := ioutil.ReadFile(path2)
	if err != nil {
		return false, err
	}
	h1 := md5.New()
	h2 := md5.New()
	h1.Write(file1)
	h2.Write(file2)

	return bytes.Compare(h1.Sum(nil), h2.Sum(nil)) == 0, nil

}

func stripDiff(path1, path2 string) (bool, error) {
	file1, err := ioutil.ReadFile(path1)
	if err != nil {
		return false, err
	}
	file2, err := ioutil.ReadFile(path2)
	if err != nil {
		return false, err
	}

	file1 = bytes.Replace(file1, []byte("\r\n"), []byte{}, -1)
	file1 = bytes.Replace(file1, []byte("\n"), []byte{}, -1)
	file1 = bytes.Replace(file1, []byte(" "), []byte{}, -1)

	file2 = bytes.Replace(file2, []byte("\r\n"), []byte{}, -1)
	file2 = bytes.Replace(file2, []byte("\n"), []byte{}, -1)
	file2 = bytes.Replace(file2, []byte(" "), []byte{}, -1)

	h1 := md5.New()
	h2 := md5.New()
	h1.Write(file1)
	h2.Write(file2)

	return bytes.Compare(h1.Sum(nil), h2.Sum(nil)) == 0, nil
}

func writeRuntimeInfoToState(state *models.Run, info lrunInfo) {
	state.Time = info.Time
	state.Memory = info.Memory
}

func Validate(state *models.Run) error {
	info, err := parseLRUN("lrun.txt")
	if err != nil {
		state.Status = RUN_TIME_ERROR
		return nil
	}
	if !validateLRUN(state, info) {
		return nil
	}

	ok, err := diff("out.txt", "user_out.txt")
	if err != nil {
		return err
	} else if ok {
		state.Status = ACCEPT
		writeRuntimeInfoToState(state, info)
		return nil
	}

	ok, err = stripDiff("out.txt", "user_out.txt")
	if err != nil {
		return err
	} else if ok {
		state.Status = PRESENTATION_ERROR
		writeRuntimeInfoToState(state, info)
		return nil
	}

	state.Status = WRONG_ANSWER
	writeRuntimeInfoToState(state, info)
	return nil
}
