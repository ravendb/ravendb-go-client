package ravendb

import (
	"fmt"
	"sync"
	"time"
)

// RangeValue represents an inclusive integer range min to max
type RangeValue struct {
	Min     int
	Max     int
	Current AtomicInteger
}

// NewRangeValue creates a new RangeValue
func NewRangeValue(min int, max int) *RangeValue {
	res := &RangeValue{
		Min: min,
		Max: max,
	}
	res.Current.Set(min - 1)
	return res
}

// HiLoIDGenerator generates document ids server side
type HiLoIDGenerator struct {
	generatorLock           sync.Mutex
	_store                  *DocumentStore
	_tag                    string
	prefix                  string
	_lastBatchSize          int
	_lastRangeDate          time.Time
	_dbName                 string
	_identityPartsSeparator string
	_range                  *RangeValue
	serverTag               string
}

// NewHiLoIdGenerator creates a HiLoIDGenerator
func NewHiLoIdGenerator(tag string, store *DocumentStore, dbName string, identityPartsSeparator string) *HiLoIDGenerator {
	return &HiLoIDGenerator{
		_store:                  store,
		_tag:                    tag,
		_dbName:                 dbName,
		_identityPartsSeparator: identityPartsSeparator,
		_range:                  NewRangeValue(1, 0),
	}
}

func (g *HiLoIDGenerator) getDocumentIDFromID(nextID int) string {
	return fmt.Sprintf("%s%d-%s", g.prefix, nextID, g.serverTag)
}

// GenerateDocumentID returns next key
func (g *HiLoIDGenerator) GenerateDocumentID(entity Object) string {
	// TODO: propagate error
	id, _ := g.nextID()
	return g.getDocumentIDFromID(id)
}

func (g *HiLoIDGenerator) nextID() (int, error) {
	for {
		// local range is not exhausted yet
		rangev := g._range
		id := rangev.Current.IncrementAndGet()
		if id <= rangev.Max {
			return id, nil
		}

		// local range is exhausted , need to get a new range
		g.generatorLock.Lock()
		defer g.generatorLock.Unlock()

		id = rangev.Current.Get()
		if id <= rangev.Max {
			return id, nil
		}
		err := g.getNextRange()
		if err != nil {
			return 0, err
		}
	}
}

func (g *HiLoIDGenerator) getNextRange() error {
	hiloCommand := NewNextHiLoCommand(g._tag, g._lastBatchSize, &g._lastRangeDate,
		g._identityPartsSeparator, g._range.Max)
	re := g._store.GetRequestExecutor()
	err := re.ExecuteCommand(hiloCommand)
	if err != nil {
		return err
	}
	result := hiloCommand.Result
	g.prefix = result.Prefix
	g.serverTag = result.ServerTag
	g._lastRangeDate = time.Time(*result.LastRangeAt)
	g._lastBatchSize = result.LastSize
	g._range = NewRangeValue(result.Low, result.High)
	return nil
}

// ReturnUnusedRange returns unused range to the server
func (g *HiLoIDGenerator) ReturnUnusedRange() error {
	returnCommand := NewHiLoReturnCommand(g._tag, g._range.Current.Get(), g._range.Max)
	re := g._store.GetRequestExecutorWithDatabase(g._dbName)
	return re.ExecuteCommand(returnCommand)
}
