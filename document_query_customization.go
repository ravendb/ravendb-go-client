package ravendb

import "time"

// Note: in Java it's hidden behind IDocumentQueryCustomization

// DocumentQueryCustomization allows customizing query
type DocumentQueryCustomization struct {
	query *AbstractDocumentQuery
}

// GetQueryOperation returns raw query operation that will be sent to the server
func (d *DocumentQueryCustomization) GetQueryOperation() *QueryOperation {
	return d.query.queryOperation
}

// AddBeforeQueryExecutedListener allows you to modify index query before it's executed
func (d *DocumentQueryCustomization) AddBeforeQueryExecutedListener(action func(*IndexQuery)) int {
	return d.query.addBeforeQueryExecutedListener(action)
}

// RemoveBeforeQueryExecutedListener removes listener added with AddBeforeQueryExecutedListener
func (d *DocumentQueryCustomization) RemoveBeforeQueryExecutedListener(idx int) {
	d.query.removeBeforeQueryExecutedListener(idx)
}

// AddAfterQueryExecutedListener adds a callback to get the results of the query
func (d *DocumentQueryCustomization) AddAfterQueryExecutedListener(action func(*QueryResult)) int {
	return d.query.addAfterQueryExecutedListener(action)
}

// RemoveAfterQueryExecutedListener removes callback added with AddAfterQueryExecutedListener
func (d *DocumentQueryCustomization) RemoveAfterQueryExecutedListener(idx int) {
	d.query.removeAfterQueryExecutedListener(idx)
}

// AddAfterStreamExecutedCallback adds a callback to get stream result
func (d *DocumentQueryCustomization) AddAfterStreamExecutedCallback(action func(map[string]interface{})) int {
	return d.query.addAfterStreamExecutedListener(action)
}

// RemoveAfterStreamExecutedCallback remove callback added with AddAfterStreamExecutedCallback
func (d *DocumentQueryCustomization) RemoveAfterStreamExecutedCallback(idx int) {
	d.query.removeAfterStreamExecutedListener(idx)
}

// NoCaching disables caching for query results
func (d *DocumentQueryCustomization) NoCaching() {
	d.query.noCaching()
}

// NoTracking disables tracking for quried entities by Raven's Unit of Work
// Using this option prevents hodling query results in memory
func (d *DocumentQueryCustomization) NoTracking() {
	d.query.noTracking()
}

// RandomOrdering orders search results randomly.
func (d *DocumentQueryCustomization) RandomOrdering() {
	d.query.randomOrdering()
}

// RandomOrdering orders search results randomly with a given seed.
// This is useful for repeatable random queries
func (d *DocumentQueryCustomization) RandomOrderingWithSeed(seed string) {
	d.query.randomOrderingWithSeed(seed)
}

// WaitForNonStaleResults instructs the query to wait for non results.
// waitTimeout of 0 means infinite timeout
// This shouldn't be used outside of unit tests  unless you are well aware of the implications
func (d *DocumentQueryCustomization) WaitForNonStaleResults(waitTimeout time.Duration) {
	d.query._waitForNonStaleResults(waitTimeout)
}
