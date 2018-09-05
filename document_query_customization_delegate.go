package ravendb

import "time"

type IDocumentQueryCustomization = DocumentQueryCustomizationDelegate

type DocumentQueryCustomizationDelegate struct {
	query *AbstractDocumentQuery
}

func NewDocumentQueryCustomizationDelegate(query *AbstractDocumentQuery) *DocumentQueryCustomizationDelegate {
	return &DocumentQueryCustomizationDelegate{
		query: query,
	}
}

func (d *DocumentQueryCustomizationDelegate) getQueryOperation() *QueryOperation {
	return d.query.GetQueryOperation()
}

func (d *DocumentQueryCustomizationDelegate) addBeforeQueryExecutedListener(action func(*IndexQuery)) *IDocumentQueryCustomization {
	d.query._addBeforeQueryExecutedListener(action)
	return d
}

func (d *DocumentQueryCustomizationDelegate) removeBeforeQueryExecutedListener(idx int) *IDocumentQueryCustomization {
	d.query._removeBeforeQueryExecutedListener(idx)
	return d
}

func (d *DocumentQueryCustomizationDelegate) addAfterQueryExecutedListener(action func(*QueryResult)) *IDocumentQueryCustomization {
	d.query._addAfterQueryExecutedListener(action)
	return d
}

func (d *DocumentQueryCustomizationDelegate) removeAfterQueryExecutedListener(idx int) *IDocumentQueryCustomization {
	d.query._removeAfterQueryExecutedListener(idx)
	return d
}

func (d *DocumentQueryCustomizationDelegate) addAfterStreamExecutedCallback(action func(ObjectNode)) *IDocumentQueryCustomization {
	d.query._addAfterStreamExecutedListener(action)
	return d
}

func (d *DocumentQueryCustomizationDelegate) removeAfterStreamExecutedCallback(idx int) *IDocumentQueryCustomization {
	d.query._removeAfterStreamExecutedListener(idx)
	return d
}

func (d *DocumentQueryCustomizationDelegate) noCaching() *IDocumentQueryCustomization {
	d.query._noCaching()
	return d
}

func (d *DocumentQueryCustomizationDelegate) noTracking() *IDocumentQueryCustomization {
	d.query._noTracking()
	return d
}

func (d *DocumentQueryCustomizationDelegate) randomOrdering() *IDocumentQueryCustomization {
	d.query._randomOrdering()
	return d
}

func (d *DocumentQueryCustomizationDelegate) randomOrderingWithSeed(seed string) *IDocumentQueryCustomization {
	d.query._randomOrderingWithSeed(seed)
	return d
}

func (d *DocumentQueryCustomizationDelegate) waitForNonStaleResults() *IDocumentQueryCustomization {
	d.query._waitForNonStaleResults(0)
	return d
}

func (d *DocumentQueryCustomizationDelegate) waitForNonStaleResultsWithTimeout(waitTimeout time.Duration) *IDocumentQueryCustomization {
	d.query._waitForNonStaleResults(waitTimeout)
	return d
}
