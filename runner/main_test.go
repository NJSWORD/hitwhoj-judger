package runner

import (
	"fmt"
	"testing"
)

func TestParseLRUN(t *testing.T) {
	info, err := parseLRUN("/tmp/lrun.txt")
	if err != nil {
		t.Fatal(err.Error())
	}
	if info.Memory != 364544 || info.Time != 1 || info.Signaled != 1 || info.ExitCode != 0 || info.TermSig != 31 || info.Exceed != "none" {
		fmt.Println(info)
		t.Fatal("not match")
	}
}

func TestDiff(t *testing.T) {
	if ok, _ := diff("/tmp/f1", "/tmp/f2"); ok {
		t.Fatal("wrong!")
	} else {
		fmt.Println(ok)
	}
}

func TestStripDiff(t *testing.T) {
	if ok, _ := stripDiff("/tmp/pe1", "/tmp/pe2"); !ok {
		t.Fatal("PE validator error")
	}
}
