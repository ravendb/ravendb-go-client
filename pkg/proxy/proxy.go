package proxy

// based on https://raw.githubusercontent.com/elazarl/goproxy/master/examples/goproxy-httpdump/httpdump.go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
)

var (
	tr                              = transport.Transport{Proxy: transport.ProxyFromEnvironment}
	proxyLogFilePath string
	proxyLogFile     *os.File
	sessionID        int32
	muLog            sync.Mutex
)

func getNextSessionID() int {
	v := atomic.AddInt32(&sessionID, 1)
	return int(v)
}

func clearSessionID() {
	atomic.StoreInt32(&sessionID, 0)
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		s := fmt.Sprintf(format, args...)
		panic(s)
	}
}

func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

func openLogFile(logPath string) {
	if logPath == "" {
		return
	}
	dir := filepath.Dir(logPath)
	err := os.MkdirAll(dir, 0755)
	must(err)
	f, err := os.Create(logPath)
	must(err)
	proxyLogFile = f
	proxyLogFilePath = logPath
	fmt.Printf("Logging to %s\n", logPath)
	clearSessionID()
}

func closeLogFile() {
	if proxyLogFile != nil {
		proxyLogFile.Close()
		proxyLogFile = nil
	}
}

// CloseLogFile closes the log file
func CloseLogFile() {
	// print buffered log
	lgShort("")

	muLog.Lock()
	defer muLog.Unlock()
	closeLogFile()
}

// ChangeLogFile changes name of log file
func ChangeLogFile(logPath string) {
	muLog.Lock()
	defer muLog.Unlock()

	if proxyLogFilePath == logPath {
		return
	}
	closeLogFile()
	openLogFile(logPath)
}

func lg(d []byte) {
	muLog.Lock()
	defer muLog.Unlock()
	if proxyLogFile != nil {
		proxyLogFile.Write(d)
		proxyLogFile.Sync()
	}
}

var (
	previousLogLine string
	logLineCount    int
)

func lgShort(s string) {
	muLog.Lock()
	defer muLog.Unlock()

	// some tests create a lot of the same requests so to eliminate the noise we delay
	// logging and combine subseqent logs that are the same into one
	if s == previousLogLine {
		logLineCount++
		return
	}
	if previousLogLine != "" {
		if logLineCount <= 1 {
			os.Stdout.WriteString(previousLogLine)
		} else {
			s := strings.TrimSpace(previousLogLine)
			s = fmt.Sprintf("%s x %d\n", s, logLineCount)
			os.Stdout.WriteString(s)
		}
	}
	previousLogLine = s
	logLineCount = 1
}

// stoppableListener serves stoppableConn and tracks their lifetime to notify
// when it is safe to terminate the application.
type stoppableListener struct {
	net.Listener
	sync.WaitGroup
}

type stoppableConn struct {
	net.Conn
	wg *sync.WaitGroup
}

func newStoppableListener(l net.Listener) *stoppableListener {
	return &stoppableListener{l, sync.WaitGroup{}}
}

func (sl *stoppableListener) Accept() (net.Conn, error) {
	c, err := sl.Listener.Accept()
	if err != nil {
		return c, err
	}
	sl.Add(1)
	return &stoppableConn{c, &sl.WaitGroup}, nil
}

func (sc *stoppableConn) Close() error {
	sc.wg.Done()
	return sc.Conn.Close()
}

// BufferCloser is a wrapper around bytes.Buffer that adds io.Close method
// to make it io.ReadCloser
type BufferCloser struct {
	*bytes.Buffer
}

// NewBufferCloser creates new BufferClose
func NewBufferCloser(buf *bytes.Buffer) *BufferCloser {
	if buf == nil {
		buf = &bytes.Buffer{}
	}
	return &BufferCloser{
		Buffer: buf,
	}
}

// Close implements io.Close interface
func (b *BufferCloser) Close() error {
	// nothing to do
	return nil
}

// SessionData has info about a request/response session
type SessionData struct {
	reqBody  bytes.Buffer
	respBody *BufferCloser
}

func NewSessionData() *SessionData {
	return &SessionData{}
}

func handleOnRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	panicIf(req == nil, "req == nil")
	sd := NewSessionData()
	ctx.UserData = sd

	if req.Body != nil {
		req.Body = ioutil.NopCloser(io.TeeReader(req.Body, &sd.reqBody))
	}
	return req, nil
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

func getRequestSummary(req *http.Request) string {
	reqURI := req.RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}
	return fmt.Sprintf("PROXY %s %s HTTP/%d.%d\r\n", valueOrDefault(req.Method, "GET"),
		reqURI, req.ProtoMajor, req.ProtoMinor)
}

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// a copy of httputil.DumpRequest that doesn't touch req.Body, which is racy sometimes with
// system code
func dumpRequest(req *http.Request) []byte {
	var b bytes.Buffer

	// By default, print out the unmodified req.RequestURI, which
	// is always set for incoming server requests. But because we
	// previously used req.URL.RequestURI and the docs weren't
	// always so clear about when to use DumpRequest vs
	// DumpRequestOut, fall back to the old way if the caller
	// provides a non-server Request.
	reqURI := req.RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}

	fmt.Fprintf(&b, "%s %s HTTP/%d.%d\r\n", valueOrDefault(req.Method, "GET"),
		reqURI, req.ProtoMajor, req.ProtoMinor)

	absRequestURI := strings.HasPrefix(req.RequestURI, "http://") || strings.HasPrefix(req.RequestURI, "https://")
	if !absRequestURI {
		host := req.Host
		if host == "" && req.URL != nil {
			host = req.URL.Host
		}
		if host != "" {
			fmt.Fprintf(&b, "Host: %s\r\n", host)
		}
	}

	if len(req.TransferEncoding) > 0 {
		fmt.Fprintf(&b, "Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
	}
	if req.Close {
		fmt.Fprintf(&b, "Connection: close\r\n")
	}

	req.Header.WriteSubset(&b, reqWriteExcludeHeaderDump)
	io.WriteString(&b, "\r\n")
	return b.Bytes()
}

func lgReq(ctx *goproxy.ProxyCtx, reqBody []byte, respBody []byte) {
	reqSummary := getRequestSummary(ctx.Req)
	lgShort(reqSummary)

	reqBody = maybePrettyPrintJSON(reqBody)
	respBody = maybePrettyPrintJSON(respBody)

	var buf bytes.Buffer
	s := fmt.Sprintf("=========== %d:\n", getNextSessionID())
	buf.WriteString(s)
	d := dumpRequest(ctx.Req)
	buf.Write(d)
	buf.Write(reqBody)

	s = "\n--------\n"
	buf.WriteString(s)
	if ctx.Resp != nil {
		d, err := httputil.DumpResponse(ctx.Resp, false)
		if err == nil {
			buf.Write(d)
		}
		buf.Write(respBody)
		buf.WriteString("\n")
	}

	lg(buf.Bytes())
}

// retruns copy of resp.Body but also makes it available for subsequent reads
func getCopyOfResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(d))
	return d, nil
}

func handleOnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	panicIf(resp != ctx.Resp, "resp != ctx.Resp")

	sd := ctx.UserData.(*SessionData)
	reqBody := sd.reqBody.Bytes()
	respBody, _ := getCopyOfResponseBody(resp)
	lgReq(ctx, reqBody, respBody)

	return resp
}

// Run starts a proxy
func Run(logPath string) {
	ChangeLogFile(logPath)

	addr := ":8888"
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	proxy.OnRequest().DoFunc(handleOnRequest)
	proxy.OnResponse().DoFunc(handleOnResponse)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen:", err)
	}

	sl := newStoppableListener(l)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Println("Got SIGINT exiting")
		sl.Add(1)
		sl.Close()
		//logger.Close()
		sl.Done()
	}()
	fmt.Printf("Starting proxy on %s\n", addr)
	http.Serve(sl, proxy)
	sl.Wait()
	log.Println("All connections closed - exit")
}
