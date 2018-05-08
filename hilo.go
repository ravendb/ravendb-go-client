package ravendb

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RangeValue represents an inclusive integer range min to max
type RangeValue struct {
	Min     int
	Max     int
	Current int
}

// NewRangeValue creates a new RangeValue
func NewRangeValue(minID int, maxID int) *RangeValue {
	return &RangeValue{
		Min:     minID,
		Max:     maxID,
		Current: minID - 1,
	}
}

// Next returns next id
func (r *RangeValue) Next() int {
	// TODO: make this atomic
	r.Current++
	return r.Current
}

// Curr returns current id
func (r *RangeValue) Curr() int {
	// TODO: make this atomic
	return r.Current
}

// NewHiLoReturnCommand creates a HiLoReturn command
func NewHiLoReturnCommand(tag string, last, end int) *RavenCommand {
	path := fmt.Sprintf("hilo/return?tag=%s&end=%d&last=%d", tag, end, last)
	url := "{url}/databases/{db}/" + path
	res := &RavenCommand{
		Method:      http.MethodPut,
		URLTemplate: url,
	}
	return res
}

// ExecuteHiLoReturnCommand executes HiLoReturnCommand
func ExecuteHiLoReturnCommand(exec CommandExecutorFunc, cmd *RavenCommand) error {
	return excuteCmdWithEmptyResult(exec, cmd)
}

const (
	// Python does "0001-01-01 00:00:00"
	// Java sends more complicated format https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/primitives/NetISO8601Utils.java#L8
	timeFormat = "2006-02-01 15:04:05"
)

// NewNextHiLoCommand creates a NextHiLoCommand
func NewNextHiLoCommand(tag string, lastBatchSize int, lastRangeAt time.Time, identityPartsSeparator string, lastRangeMax int) *RavenCommand {
	lastRangeAtStr := quoteKey(lastRangeAt.Format(timeFormat))
	path := fmt.Sprintf("hilo/next?tag=%s&lastBatchSize=%dd&lastRangeAt=%d&identityPartsSeparator=%s&lastMax=%d", tag, lastBatchSize, lastRangeAtStr, identityPartsSeparator, lastRangeMax)
	url := "{url}/databases/{db}/" + path
	res := &RavenCommand{
		Method:      http.MethodGet,
		URLTemplate: url,
	}
	return res
}

// NextHiLoResult is a result of NextHiLoResult command
type NextHiLoResult struct {
	Prefix      string `json:"Prefix"`
	Low         int    `json:"Low"`
	High        int    `json:"High"`
	LastSize    int    `json:"LastSize"`
	ServerTag   string `json:"ServerTag"`
	LastRangeAt string `json:"LastRangeAt"`
}

const (
	// time format returned by the server
	// 2018-05-08T05:20:31.5233900Z
	serverTimeFormat = "2006-01-02T15:04:05.999999999Z"
)

// GetLastRangeAt parses LastRangeAt which is in a format:
// 2018-05-08T05:20:31.5233900Z
func (r *NextHiLoResult) GetLastRangeAt() time.Time {
	t, err := time.Parse(serverTimeFormat, r.LastRangeAt)
	must(err) // TODO: should silently fail? return an error?
	return t
}

// ExecuteNewNextHiLoCommand executes NextHiLoResult command
func ExecuteNewNextHiLoCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*NextHiLoResult, error) {
	return nil, nil
}

// HiLoKeyGenerator generates keys server side
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/hilo/hilo_generator.py#L63
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/identity/HiLoIdGenerator.java#L14
type HiLoKeyGenerator struct {
	tag                    string
	store                  *DocumentStore
	dbName                 string
	lastRangeAt            time.Time
	lastBatchSize          int
	rangev                 *RangeValue
	prefix                 string
	serverTag              string
	convetions             *DocumentConventions
	identityPartsSeparator string
	lock                   sync.Mutex
}

// NewHiLoKeyGenerator creates a HiLoKeyGenerator
func NewHiLoKeyGenerator(tag string, store *DocumentStore, dbName string) *HiLoKeyGenerator {
	if dbName == "" {
		dbName = store.database
	}
	res := &HiLoKeyGenerator{
		tag:           tag,
		store:         store,
		dbName:        dbName,
		lastRangeAt:   time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		lastBatchSize: 0,
		rangev:        NewRangeValue(1, 0),
		prefix:        "",
		serverTag:     "",
		convetions:    store.Conventions,
	}
	res.identityPartsSeparator = res.convetions.IdentityPartsSeparator
	return res
}

// GetDocumentKeyFromID creates key from id
func (g *HiLoKeyGenerator) GetDocumentKeyFromID(nextID int) string {
	return fmt.Sprintf("%s%d-%s", g.prefix, nextID, g.serverTag)
}

// GenerateDocumentKey returns next key
func (g *HiLoKeyGenerator) GenerateDocumentKey() int {
	for {
		// local range is not exhausted yet
		rangev := g.rangev

		id := rangev.Next()
		if id <= rangev.Max {
			return id
		}

		// local range is exhausted , need to get a new range
		g.lock.Lock()
		id = rangev.Curr()
		if id <= rangev.Max {
			return id
		}

		g.getNextRange()
		g.lock.Unlock()
	}
}

func (g *HiLoKeyGenerator) getNextRange() {
	cmd := NewNextHiLoCommand(g.tag, g.lastBatchSize, g.lastRangeAt,
		g.identityPartsSeparator, g.rangev.Max)
	// TODO: use store.getRequestsExecutor().Exec()
	exec := g.store.getSimpleExecutor()
	// TODO: propagate the error
	res, _ := ExecuteNewNextHiLoCommand(exec, cmd)
	g.prefix = res.Prefix
	g.serverTag = res.ServerTag
	g.lastRangeAt = res.GetLastRangeAt()
	g.lastBatchSize = res.LastSize
	g.rangev = NewRangeValue(res.Low, res.High)
}

func (g *HiLoKeyGenerator) returnUnusedRange() error {
	cmd := NewHiLoReturnCommand(g.tag, g.rangev.Curr(), g.rangev.Max)
	// TODO: use store.getRequestsExecutor().Exec()
	exec := g.store.getSimpleExecutor()
	err := ExecuteHiLoReturnCommand(exec, cmd)
	return err
}

// TODO:
// * MultiDatabaseHiLoKeyGenerator
// * MultiTypeHiLoKeyGenerator
