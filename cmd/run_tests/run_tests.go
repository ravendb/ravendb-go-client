package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func runProxyProcess() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/loggingproxy/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	must(err)
	return cmd
}

func runJava() {
	logFileTmpl := "trace_hilo_java.txt"
	go runProxy(logFileTmpl)
	defer closeProxyLogFile()

	//cmdProxy := runProxy()
	//defer cmdProxy.Process.Kill()

	// Running just one maven test: https://stackoverflow.com/a/18136440/2898
	// mvn -Dtest=HiLoTest test
	// mvn test
	cmd := exec.Command("mvn", "-Dtest=HiLoTest", "test")
	cmd.Dir = path.Join("..", "ravendb-jvm-client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	must(err)
	cmd.Wait()
}

func runGo() {
	cmdProxy := runProxyProcess()
	defer cmdProxy.Process.Kill()

	cmd := exec.Command("go", "test", "-race")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	must(err)
	cmd.Wait()
}

func main() {
	var arg string
	if len(os.Args) == 2 {
		arg = os.Args[1]
	}

	switch arg {
	case "-java":
		runJava()
	case "-go":
		runGo()
	default:
		fmt.Printf("Needs to privide an argument -java or -go\n")
		os.Exit(1)
	}
}
