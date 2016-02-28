package languages

type language struct {
	SourceFile   string
	CompileCmd   string
	RunCmd       string
	Syscalls     string
	TimeOffset   float32
	MemoryOffset float32
}

var Languages map[string]language

// 部分来自 https://github.com/QingdaoU/OnlineJudge/blob/master/judge/language.py
// Dirty Code!
// CompileCmd 需要 fmt.Sprintf(CompileCmd, workDir, workDir) (分别代表输出地址，输入地址)
// RunCmd 需要 fmt.Sprintf(RunCmd, workDir) (程序地址)
func init() {
	Languages = make(map[string]language)

	Languages["c++"] = language{
		"Main.cpp",
		" g++ -std=c99 -O2 -o %s/Main -DONLINEJUDGE -w -pipe -lm -fomit-frame-pointer %s/Main.cpp ",
		" %s/Main ",
		"!execve:k,flock:k,ptrace:k,sync:k,fdatasync:k,fsync:k,msync,sync_file_range:k,syncfs:k,unshare:k,setns:k,clone:k,query_module:k,sysinfo:k,syslog:k,sysfs:k",
		1,
		1,
	}

	Languages["c"] = language{
		"Main.c",
		" gcc -std=c++0x -O2 -o %s/Main -DONLINEJUDGE -w -pipe -lm -fomit-frame-pointer %s/Main.c ",
		" %s/Main ",
		"!execve:k,flock:k,ptrace:k,sync:k,fdatasync:k,fsync:k,msync,sync_file_range:k,syncfs:k,unshare:k,setns:k,clone:k,query_module:k,sysinfo:k,syslog:k,sysfs:k",
		1,
		1,
	}

	Languages["java"] = language{
		"Main.java",
		" javac -d %s %s/Main.java ", // javac -g:none -Xlint Main.java
		" java %s/Main ",             //  java -client Main
		"!execve:k,flock:k,ptrace:k,sync:k,fdatasync:k,fsync:k,msync,sync_file_range:k,syncfs:k,unshare:k,setns:k,clone[a&268435456==268435456]:k,query_module:k,sysinfo:k,syslog:k,sysfs:k",
		2,
		2,
	}
}
