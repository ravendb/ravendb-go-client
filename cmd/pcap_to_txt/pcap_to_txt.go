package main

// go run cmd/pcap_to_txt/pcap_to_txt.go logs/trace_indexes_from_client_go.pcap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/ga0/netgraph/ngnet"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

// This program converts a .pcap file into a text file that
// shows http requests and responses

func usageAndExit() {
	exe := filepath.Base(os.Args[0])
	fmt.Printf("%s: ${in.pcap} [${out.txt}]\n", exe)
	os.Exit(1)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	file io.Writer
)

func isNL(c byte) bool {
	return c == 0xd || c == 0xa
}

func isBinaryData(d []byte) bool {
	for _, b := range d {
		if b < 32 && !isNL(b) {
			return true
		}
	}
	return false
}

func asHex(d []byte) ([]byte, bool) {
	if !isBinaryData(d) {
		return d, false
	}
	if len(d) > 32 {
		d = d[:32]
	}
	s := ""
	for i, b := range d {
		if i > 0 && i%16 == 0 {
			s += "\n"
		}
		s += fmt.Sprintf("%02x ", b)
	}
	return []byte(s), true
}

// if d is a valid json, pretty-print it
func prettyPrintMaybeJSON(d []byte) []byte {
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
	d := prettyPrintMaybeJSON(req.Body)
	fmt.Fprintf(file, "%s\n", d)
}

func printHTTPResponseEvent(resp *ngnet.HTTPResponseEvent, no int) {
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
	d := prettyPrintMaybeJSON(resp.Body)
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

func dumpRequests() {
	sort.Slice(requests, func(i, j int) bool {
		t1 := requests[i].req.HTTPEvent.Start
		t2 := requests[j].req.HTTPEvent.Start
		return t1.After(t2)
	})

	for n, rr := range requests {
		printHTTPRequestEvent(rr.req, n+1)
		printHTTPResponseEvent(rr.rsp, n+1)
	}

	if len(unmatchedRequests) > 0 {
		fmt.Printf("%d unmatched requests\n", len(unmatchedRequests))
		for n, r := range unmatchedRequests {
			printHTTPRequestEvent(r, n)
		}
	}

	if len(unmatchedResponses) > 0 {
		fmt.Printf("%d unmatched responses\n", len(unmatchedResponses))
		for n, r := range unmatchedResponses {
			printHTTPResponseEvent(r, n)
		}
	}
}

func rememberRequest(r *ngnet.HTTPRequestEvent) {
	if len(unmatchedResponses) == 0 {
		unmatchedRequests = append(unmatchedRequests, r)
		return
	}
	for idx, rsp := range unmatchedResponses {
		if r.HTTPEvent.StreamSeq == rsp.HTTPEvent.StreamSeq {
			rr := reqRsp{
				req: r,
				rsp: rsp,
			}
			requests = append(requests, rr)
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
			rr := reqRsp{
				req: req,
				rsp: r,
			}
			requests = append(requests, rr)
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

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		usageAndExit()
	}
	pcapPath := args[0]
	fmt.Printf("Started on %s\n", pcapPath)
	file = os.Stdout
	dumpHTTPFromPcap(pcapPath)
}
