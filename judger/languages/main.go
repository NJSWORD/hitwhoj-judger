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

func init() {
	Languages = make(map[string]language)

	Languages["c++"] = language{
		"Main.cpp",
		" g++ -std=c99 -O2 -oMain -DONLINEJUDGE -w -pipe -lm -fomit-frame-pointer Main.cpp ",
		" ./Main ",
		"!execve:k,flock:k,ptrace:k,sync:k,fdatasync:k,fsync:k,msync,sync_file_range:k,syncfs:k,unshare:k,setns:k,clone:k,query_module:k,sysinfo:k,syslog:k,sysfs:k",
		1,
		1,
	}

	Languages["c"] = language{
		"Main.c",
		" gcc -std=c++0x -O2 -oMain -DONLINEJUDGE -w -pipe -lm -fomit-frame-pointer Main.c ",
		" ./Main ",
		"!execve:k,flock:k,ptrace:k,sync:k,fdatasync:k,fsync:k,msync,sync_file_range:k,syncfs:k,unshare:k,setns:k,clone:k,query_module:k,sysinfo:k,syslog:k,sysfs:k",
		1,
		1,
	}

	Languages["java"] = language{
		"Main.java",
		" javac Main.java ", // javac -g:none -Xlint Main.java
		" java Main ", //  java -client Main
		"!execve:k,flock:k,ptrace:k,sync:k,fdatasync:k,fsync:k,msync,sync_file_range:k,syncfs:k,unshare:k,setns:k,clone[a&268435456==268435456]:k,query_module:k,sysinfo:k,syslog:k,sysfs:k",
		2,
		2,
	}
}
