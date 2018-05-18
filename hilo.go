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
func NewRangeValue(min int, max int) *RangeValue {
	return &RangeValue{
		Min:     min,
		Max:     max,
		Current: min - 1,
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
	path := fmt.Sprintf("hilo/next?tag=%s&lastBatchSize=%d&lastRangeAt=%s&identityPartsSeparator=%s&lastMax=%d", tag, lastBatchSize, lastRangeAtStr, identityPartsSeparator, lastRangeMax)
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
	var res NextHiLoResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
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
func (g *HiLoKeyGenerator) GenerateDocumentKey() string {
	// TODO: propagate error
	id, _ := g.nextID()
	return g.GetDocumentKeyFromID(id)
}

func (g *HiLoKeyGenerator) nextID() (int, error) {
	g.lock.Lock()
	defer g.lock.Unlock()
	for {
		// local range is not exhausted yet
		rangev := g.rangev
		id := rangev.Next()
		if id <= rangev.Max {
			return id, nil
		}

		// local range is exhausted , need to get a new range
		err := g.getNextRange()
		if err != nil {
			return 0, err
		}
	}
}

func (g *HiLoKeyGenerator) getNextRange() error {
	exec := g.store.GetRequestExecutor("").GetCommandExecutor(false)
	cmd := NewNextHiLoCommand(g.tag, g.lastBatchSize, g.lastRangeAt,
		g.identityPartsSeparator, g.rangev.Max)
	res, err := ExecuteNewNextHiLoCommand(exec, cmd)
	if err != nil {
		return err
	}
	g.prefix = res.Prefix
	g.serverTag = res.ServerTag
	g.lastRangeAt = res.GetLastRangeAt()
	g.lastBatchSize = res.LastSize
	g.rangev = NewRangeValue(res.Low, res.High)
	return nil
}

// ReturnUnusedRange returns unused range
func (g *HiLoKeyGenerator) ReturnUnusedRange() error {
	cmd := NewHiLoReturnCommand(g.tag, g.rangev.Curr(), g.rangev.Max)
	// TODO: use store.getRequestsExecutor().Exec()
	exec := g.store.getSimpleExecutor()
	return ExecuteHiLoReturnCommand(exec, cmd)
}

// MultiTypeHiLoKeyGenerator manages per-type HiLoKeyGenerator
type MultiTypeHiLoKeyGenerator struct {
	store  *DocumentStore
	dbName string
	// maps type name to its generator
	keyGeneratorsByTag map[string]*HiLoKeyGenerator
	lock               sync.Mutex // protects keyGeneratorsByTag
}

// NewMultiTypeHiLoKeyGenerator creates MultiTypeHiLoKeyGenerator
func NewMultiTypeHiLoKeyGenerator(store *DocumentStore, dbName string) *MultiTypeHiLoKeyGenerator {
	return &MultiTypeHiLoKeyGenerator{
		store:              store,
		dbName:             dbName,
		keyGeneratorsByTag: map[string]*HiLoKeyGenerator{},
	}
}

// GenerateDocumentKey generates a unique key for entity using its type to
// partition keys
func (g *MultiTypeHiLoKeyGenerator) GenerateDocumentKey(entity interface{}) string {
	tag := defaultTransformTypeTagName(getShortTypeName(entity))
	g.lock.Lock()
	generator, ok := g.keyGeneratorsByTag[tag]
	if !ok {
		generator = NewHiLoKeyGenerator(tag, g.store, g.dbName)
		g.keyGeneratorsByTag[tag] = generator
	}
	g.lock.Unlock()
	return generator.GenerateDocumentKey()
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiTypeHiLoKeyGenerator) ReturnUnusedRange() {
	for _, generator := range g.keyGeneratorsByTag {
		generator.ReturnUnusedRange()
	}
}

// MultiDatabaseHiLoKeyGenerator manages per-database HiLoKeyGenerotr
type MultiDatabaseHiLoKeyGenerator struct {
	store      *DocumentStore
	generators map[string]*MultiTypeHiLoKeyGenerator
}

// NewMultiDatabaseHiLoKeyGenerator creates new MultiDatabaseHiLoKeyGenerator
func NewMultiDatabaseHiLoKeyGenerator(store *DocumentStore) *MultiDatabaseHiLoKeyGenerator {
	return &MultiDatabaseHiLoKeyGenerator{
		store:      store,
		generators: map[string]*MultiTypeHiLoKeyGenerator{},
	}
}

// GenerateDocumentKey generates
func (g *MultiDatabaseHiLoKeyGenerator) GenerateDocumentKey(dbName string, entity interface{}) string {
	if dbName == "" {
		dbName = g.store.database
	}
	panicIf(dbName == "", "expected non-empty dbName")
	generator, ok := g.generators[dbName]
	if !ok {
		generator = NewMultiTypeHiLoKeyGenerator(g.store, dbName)
		g.generators[dbName] = generator
	}
	return generator.GenerateDocumentKey(entity)
}

// ReturnUnusedRange returns unused range for all generators
func (g *MultiDatabaseHiLoKeyGenerator) ReturnUnusedRange() {
	for _, generator := range g.generators {
		generator.ReturnUnusedRange()
	}
}