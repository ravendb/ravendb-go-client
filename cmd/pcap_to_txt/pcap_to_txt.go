package main

// go run cmd/pcap_to_txt/pcap_to_txt.go logs/trace_indexes_from_client_go.pcap

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ga0/netgraph/ngnet"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

// This program converts a .pcap file into a text file that
// shows http requests and responses

var (
	flgStdout bool

	srcPath string
	file    io.Writer
)

func parseCmdLineArgs() {
	flag.BoolVar(&flgStdout, "stdout", false, "if true, prints to stdout instead of a file")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		usageAndExit()
	}
	srcPath = args[0]
}

func usageAndExit() {
	exe := filepath.Base(os.Args[0])
	fmt.Printf("%s: in.pcap | dir [-stdout]\n", exe)
	fmt.Print(`  converts in.pcap to in.txt
  if arg is a directory, converts all .pcap files to corresponding .txt files
  if -stdout is given, will print text version to stdout instead of writing to .txt file
`)
	os.Exit(1)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func isUnprintable(c byte) bool {
	if c < 32 {
		// 9 - tab, 10 - LF, 13 - CR
		if c == 9 || c == 10 || c == 13 {
			return false
		}
		return true
	}
	return c >= 127
}

func isBinaryData(d []byte) bool {
	for _, b := range d {
		if isUnprintable(b) {
			return true
		}
	}
	return false
}

func asHex(d []byte) ([]byte, bool) {
	if !isBinaryData(d) {
		return d, false
	}

	// convert unprintable characters to hex
	var res []byte
	for i, c := range d {
		if i > 2048 {
			break
		}
		if isUnprintable(c) {
			s := fmt.Sprintf("x%02x ", c)
			res = append(res, s...)
		} else {
			res = append(res, c)
		}
	}
	return res, true
}

// if d is a valid json, pretty-print it
func maybePrettyPrintJSON(d []byte) []byte {
	if d2, ok := asHex(d); ok {
		return d2
	}
	var m map[string]interface{}
	err := json.Unmarshal(d, &m)
	if err != nil {
		return d
	}
	d2, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return d
	}
	return d2
}

var (
	mu sync.Mutex
)

func printHTTPRequestEvent(req *ngnet.HTTPRequestEvent, no int) {
	if req == nil {
		fmt.Fprintf(file, "> Request %d missing\n", no)
		return
	}

	fmt.Fprintf(file, "> Request %d %d %s->%s\n", no, req.StreamSeq, req.ClientAddr, req.ServerAddr)
	fmt.Fprintf(file, "%s %s %s\n", req.Method, req.URI, req.Version)
	for _, h := range req.Headers {
		fmt.Fprintf(file, "%s: %s\n", h.Name, h.Value)
	}

	n := len(req.Body)
	if n == 0 {
		fmt.Fprint(file, "\n0 bytes sent\n")
		return
	}

	fmt.Fprintf(file, "\n%d bytes:\n", n)
	d := maybePrettyPrintJSON(req.Body)
	fmt.Fprintf(file, "%s\n", d)
}

func printHTTPResponseEvent(resp *ngnet.HTTPResponseEvent, no int) {
	if resp == nil {
		fmt.Fprintf(file, "< Response %d missing\n", no)
		return

	}
	fmt.Fprintf(file, "< Response %d %d %s<-%s\n", no, resp.StreamSeq, resp.ClientAddr, resp.ServerAddr)
	fmt.Fprintf(file, "%s %d %s\n", resp.Version, resp.Code, resp.Reason)
	for _, h := range resp.Headers {
		fmt.Fprintf(file, "%s: %s\n", h.Name, h.Value)
	}

	n := len(resp.Body)
	if n == 0 {
		fmt.Fprint(file, "\n0 bytes received\n")
		return
	}

	fmt.Fprintf(file, "\n%d bytes:\n", n)
	d := maybePrettyPrintJSON(resp.Body)
	fmt.Fprintf(file, "%s\n", d)
}

type reqRsp struct {
	req *ngnet.HTTPRequestEvent
	rsp *ngnet.HTTPResponseEvent
}

var (
	// hae to queue them to match requests with responses
	requests           []reqRsp
	unmatchedRequests  []*ngnet.HTTPRequestEvent
	unmatchedResponses []*ngnet.HTTPResponseEvent
)

func getRequestTime(rr reqRsp, timeDefault time.Time) time.Time {
	req := rr.req
	if req == nil {
		return rr.rsp.HTTPEvent.Start
	}
	return req.HTTPEvent.Start
}

func addRequest(req *ngnet.HTTPRequestEvent, rsp *ngnet.HTTPResponseEvent) {
	rr := reqRsp{
		req: req,
		rsp: rsp,
	}
	requests = append(requests, rr)
}

func dumpRequests() {
	if len(unmatchedRequests) > 0 {
		fmt.Printf("%d unmatched requests\n", len(unmatchedRequests))
		for _, r := range unmatchedRequests {
			addRequest(r, nil)
		}
	}
	if len(unmatchedResponses) > 0 {
		fmt.Printf("%d unmatched responses\n", len(unmatchedResponses))
		for _, r := range unmatchedResponses {
			addRequest(nil, r)
		}
	}

	t := time.Now()
	sort.Slice(requests, func(i, j int) bool {
		t1 := getRequestTime(requests[i], t)
		t2 := getRequestTime(requests[j], t)
		return t1.After(t2)
	})

	for n, rr := range requests {
		printHTTPRequestEvent(rr.req, n+1)
		printHTTPResponseEvent(rr.rsp, n+1)
	}
}

func rememberRequest(r *ngnet.HTTPRequestEvent) {
	if len(unmatchedResponses) == 0 {
		unmatchedRequests = append(unmatchedRequests, r)
		return
	}
	for idx, rsp := range unmatchedResponses {
		if r.HTTPEvent.StreamSeq == rsp.HTTPEvent.StreamSeq {
			addRequest(r, rsp)
			// remove element at idx
			unmatchedResponses = append(unmatchedResponses[:idx], unmatchedResponses[idx+1:]...)
			return
		}
	}
	unmatchedRequests = append(unmatchedRequests, r)
}

func rememberResponse(r *ngnet.HTTPResponseEvent) {
	if len(unmatchedRequests) == 0 {
		unmatchedResponses = append(unmatchedResponses, r)
		return
	}
	for idx, req := range unmatchedRequests {
		if r.HTTPEvent.StreamSeq == req.HTTPEvent.StreamSeq {
			addRequest(req, r)
			// remove element at idx
			unmatchedRequests = append(unmatchedRequests[:idx], unmatchedRequests[idx+1:]...)
			return
		}
	}
	unmatchedResponses = append(unmatchedResponses, r)
}

func runEvents(eventChan <-chan interface{}) {
	for e := range eventChan {
		switch v := e.(type) {
		case ngnet.HTTPRequestEvent:
			rememberRequest(&v)
		case ngnet.HTTPResponseEvent:
			rememberResponse(&v)
		default:
			panic(fmt.Sprintf("Unsupported event %T", e))
		}
	}
}

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct{}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	stream := &httpStream{
		net:       net,
		transport: transport,
		r:         newTCPStream(),
	}
	return stream.r
}

// httpStream will handle the actual decoding of http requests.
type httpStream struct {
	net, transport gopacket.Flow
	r              *tcpStream
}

type tcpStream struct {
	nPackets int
	buf      *bytes.Buffer
}

func newTCPStream() *tcpStream {
	return &tcpStream{
		buf: bytes.NewBuffer(nil),
	}
}

func (r *tcpStream) Reassembled(reassembly []tcpassembly.Reassembly) {
	for _, re := range reassembly {
		r.buf.Write(re.Bytes)
		r.nPackets++
	}
}

// ReassemblyComplete implements tcpassembly.Stream's ReassemblyComplete function.
func (r *tcpStream) ReassemblyComplete() {
	d := r.buf.Bytes()
	fmt.Printf("Finished reassembly. %d packets, %d bytes\n", r.nPackets, len(d))
	fmt.Printf("-----\n%s\n------\n", d)
}

func readAllPackets(packetSource *gopacket.PacketSource, assembler *tcpassembly.Assembler) {
	for packet := range packetSource.Packets() {
		netLayer := packet.NetworkLayer()
		if netLayer == nil {
			continue
		}
		transLayer := packet.TransportLayer()
		if transLayer == nil {
			continue
		}
		tcp, _ := transLayer.(*layers.TCP)
		if tcp == nil {
			continue
		}
		assembler.AssembleWithTimestamp(
			netLayer.NetworkFlow(),
			tcp,
			packet.Metadata().CaptureInfo.Timestamp)
	}
	assembler.FlushAll()
}

func dumpHTTPFromPcap(pcapPath string) {
	eventChan := make(chan interface{}, 1024)

	go runEvents(eventChan)

	handle, err := pcap.OpenOffline(pcapPath)
	panicIfErr(err)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	streamFactory := ngnet.NewHTTPStreamFactory(eventChan)
	pool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(pool)
	readAllPackets(packetSource, assembler)

	streamFactory.Wait()
	close(eventChan)
	dumpRequests()
}

func dumpHTTPFromPcap2(pcapPath string) {
	handle, err := pcap.OpenOffline(pcapPath)
	panicIfErr(err)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	streamFactory := &httpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	readAllPackets(packetSource, assembler)
	dumpRequests()
}

func isPcapFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".pcap"
}

func convertDir(dir string) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("ioutil.ReadDir('%s') failed with %s\n", dir, err)
		return
	}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			continue
		}
		if !isPcapFile(fi.Name()) {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		convertFile(path)
	}
}

// convert "foo.pcap" => "foo.txt"
func convertFilePath(s string) string {
	ext := filepath.Ext(s)
	s = s[0 : len(s)-len(ext)]
	return s + ".txt"
}

func convertFile(path string) {
	if flgStdout {
		fmt.Printf("Printing %s to stdout\n", path)
		file = os.Stdout
	} else {
		dstPath := convertFilePath(path)
		fmt.Printf("Converting %s => %s\n", path, dstPath)
		dstFile, err := os.Create(dstPath)
		if err != nil {
			fmt.Printf("os.Create('%s') failed with %s\n", dstPath, err)
		}
		defer dstFile.Close()
		file = dstFile
	}
	dumpHTTPFromPcap(path)
}

func main() {
	parseCmdLineArgs()

	st, err := os.Stat(srcPath)
	if err != nil {
		fmt.Printf("os.Stat('%s') failed with %s\n", srcPath, err)
		return
	}

	if st.IsDir() {
		if flgStdout {
			flgStdout = false
			fmt.Printf("Warning: -stdout is not supported when converting a directory\n")
		}
		convertDir(srcPath)
	} else {
		convertFile(srcPath)
	}
}
