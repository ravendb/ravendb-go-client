package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ravendb/ravendb-go-client/pkg/capture"
)

var (
	flgAddr               string
	flgPcapFile           string
	flgShowRequestSummary bool
)

func parseCmdLineArgs() {
	flag.StringVar(&flgAddr, "addr", "", "ip address for which to sniff the traffic e.g. 127.0.0.1:5383")
	flag.BoolVar(&flgShowRequestSummary, "show-request-summary", false, "if true, shows summary of requests")
	flag.StringVar(&flgPcapFile, "pcap", "", ".pcap file where we write captured packets")
	flag.Parse()
	if flgAddr == "" || flgPcapFile == "" {
		flag.Usage()
	}
}

func exitIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func createDirForFile(path string) {
	dir := filepath.Dir(path)
	if dir == "" {
		return
	}
	err := os.MkdirAll(dir, 0755)
	exitIfErr(err)
}

func main() {
	parseCmdLineArgs()
	if flgShowRequestSummary {
		capture.RequestsSummaryWriter = os.Stdout
	}
	createDirForFile(flgPcapFile)
	fmt.Printf("Capturing packets from %s to %s\n", flgAddr, flgPcapFile)
	capturer, err := capture.StartCapture(flgAddr, flgPcapFile)
	exitIfErr(err)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("Recieved signal '%s'\n", sig)
	capturer.Close()
}
