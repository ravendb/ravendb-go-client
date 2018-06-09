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

func runSingleJavaTest(className string, logFileTmpl string) {
	go proxy.Run(logFileTmpl)
	defer proxy.CloseLogFile()

	// Running just one maven test: https://stackoverflow.com/a/18136440/2898
	// mvn -Dtest=HiLoTest test
	cmd := exec.Command("mvn", fmt.Sprintf("-Dtest=%s", className), "test")
	cmd.Dir = path.Join("..", "ravendb-jvm-client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	must(err)
	err = cmd.Wait()
	must(err)
}

func runJava() {
	// TODO: for some reason when we run more than one in a sequence,
	// the second one fails. Possibly because the server fails
	// to start the second time

	//runSingleJavaTest("HiLoTest", "trace_hilo_java.txt")
	//runSingleJavaTest("GetTopologyTest", "trace_get_topology_java.txt")
	runSingleJavaTest("GetTcpInfoTest", "trace_get_tcp_info_java.txt")
}

func main() {
	var arg string
	if len(os.Args) == 2 {
		arg = os.Args[1]
	}

	switch arg {
	case "-java":
		runJava()
	default:
		fmt.Printf("Needs to privide an argument -java or -go\n")
		os.Exit(1)
	}
}
