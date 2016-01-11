package languages

type language struct {
	SourceFile string
	CompileCmd []string
}


var Languages map[string]language

func init() {
	Languages = make(map[string]language)

	Languages["c++"] = language{"Main.cpp",
		[]string{"g++", "-O2", "-oMain", "-DONLINEJUDGE", "-w", "Main.cpp"},
	}

	Languages["c"] = language{
		"Main.c",
		[]string{"gcc", "-O2", "-oMain", "-DONLINEJUDGE", "-w", "Main.c"},
	}

	// TODO: support Java
	Languages["java"] = language{
		"Main.java",
		[]string{"ls"},
	}
}
