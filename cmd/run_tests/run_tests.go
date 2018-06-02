package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
)

const (
	proxyURL = "http://localhost:8888"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func runJava() {
	logFileTmpl := "trace_hilo_java.txt"
	go proxy.Run(logFileTmpl)
	defer proxy.CloseLogFile()

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
	logFileTmpl := "trace_hilo_go.txt"
	go proxy.Run(logFileTmpl)
	defer proxy.CloseLogFile()

	cmd := exec.Command("go", "test", "-race")
	// this tells http client to use a proxy
	// https://stackoverflow.com/questions/14661511/setting-up-proxy-for-http-client
	proxyEnv := "HTTP_PROXY=" + proxyURL
	cmd.Env = append(os.Environ(), proxyEnv)

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
