package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	unlikelySep = "\x02\x01\x03"
)

func stringArrayCopy(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	return append([]string{}, a...)
}

func stringArrayContains(a []string, s string) bool {
	for _, el := range a {
		if el == s {
			return true
		}
	}
	return false
}

// equivalent of Java's containsSequence http://joel-costigliola.github.io/assertj/core/api/org/assertj/core/api/ListAssert.html#containsSequence(ELEMENT...)
// checks if a1 contains sub-sequence a2
func stringArrayContainsSequence(a1, a2 []string) bool {
	// TODO: technically it's possible for this to have false positive
	// but it's very unlikely
	s1 := strings.Join(a1, unlikelySep)
	s2 := strings.Join(a2, unlikelySep)
	return strings.Contains(s1, s2)
}

func stringArrayContainsExactly(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, s := range a1 {
		if s != a2[i] {
			return false
		}
	}
	return true
}

// stringArrayEq returns true if arrays have the same content, ignoring order
func stringArrayEq(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	if len(a1) == 0 {
		return true
	}
	a1c := stringArrayCopy(a1)
	a2c := stringArrayCopy(a2)
	sort.Strings(a1c)
	sort.Strings(a2c)
	for i, s := range a1c {
		if s != a2c[i] {
			return false
		}
	}
	return true
}

func stringArrayReverse(a []string) {
	n := len(a)
	for i := 0; i < n/2; i++ {
		a[i], a[n-1-i] = a[n-1-i], a[i]
	}
}

func int64ArrayHasDuplicates(a []int64) bool {
	if len(a) == 0 {
		return false
	}
	m := map[int64]int{}
	for _, i := range a {
		m[i]++
		if m[i] > 1 {
			return true
		}
	}
	return false
}

func jsonGetAsText(doc map[string]interface{}, key string) (string, bool) {
	v, ok := doc[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func objectNodeFieldNames(js map[string]interface{}) []string {
	var res []string
	for k := range js {
		res = append(res, k)
	}
	return res
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
// only used for debugging
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

func fileExists(path string) bool {
	st, err := os.Lstat(path)
	return err == nil && !st.IsDir()
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// can be used for http.Get() requests with better timeouts. New one must be created
// for each Get() request
func newTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial:  timeoutDialer(connectTimeout, readWriteTimeout),
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func downloadURL(url string) ([]byte, error) {
	// default timeout for http.Get() is really long, so dial it down
	// for both connection and read/write timeouts
	timeoutClient := newTimeoutClient(time.Second*120, time.Second*120)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := timeoutClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("'%s': status code not 200 (%d)", url, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func HttpDl(url string, destPath string) error {
	d, err := downloadURL(url)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(destPath, d, 0755)
}
