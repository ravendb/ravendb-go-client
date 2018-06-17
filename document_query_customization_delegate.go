package ravendb

import "time"

type DocumentQueryCustomizationDelegate struct {
	query *AbstractDocumentQuery
}

func NewDocumentQueryCustomizationDelegate(query *AbstractDocumentQuery) *DocumentQueryCustomizationDelegate {
	return &DocumentQueryCustomizationDelegate{
		query: query,
	}
}

func (d *DocumentQueryCustomizationDelegate) getQueryOperation() *QueryOperation {
	//return query.getQueryOperation();
	return nil
}

func (d *DocumentQueryCustomizationDelegate) addBeforeQueryExecutedListener(action ConsumerOfIndexQuery) *IDocumentQueryCustomization {
	//query._addBeforeQueryExecutedListener(action);
	return d
}

func (d *DocumentQueryCustomizationDelegate) removeBeforeQueryExecutedListener(action ConsumerOfIndexQuery) *IDocumentQueryCustomization {
	//query._removeBeforeQueryExecutedListener(action);
	return d
}

func (d *DocumentQueryCustomizationDelegate) addAfterQueryExecutedListener(action ConsumerOfQueryResult) *IDocumentQueryCustomization {
	//query._addAfterQueryExecutedListener(action);
	return d
}

func (d *DocumentQueryCustomizationDelegate) removeAfterQueryExecutedListener(action ConsumerOfQueryResult) *IDocumentQueryCustomization {
	//query._removeAfterQueryExecutedListener(action);
	return d
}

func (d *DocumentQueryCustomizationDelegate) addAfterStreamExecutedCallback(action ConsumerOfObjectNode) *IDocumentQueryCustomization {
	//query._addAfterStreamExecutedListener(action);
	return d
}

func (d *DocumentQueryCustomizationDelegate) removeAfterStreamExecutedCallback(action ConsumerOfObjectNode) *IDocumentQueryCustomization {
	//query._removeAfterStreamExecutedListener(action);
	return d
}

func (d *DocumentQueryCustomizationDelegate) noCaching() *IDocumentQueryCustomization {
	//query._noCaching();
	return d
}

func (d *DocumentQueryCustomizationDelegate) noTracking() *IDocumentQueryCustomization {
	//query._noTracking();
	return d
}

func (d *DocumentQueryCustomizationDelegate) randomOrdering() *IDocumentQueryCustomization {
	//query._randomOrdering();
	return d
}

func (d *DocumentQueryCustomizationDelegate) randomOrderingWithSeed(seed string) *IDocumentQueryCustomization {
	//query._randomOrdering(seed);
	return d
}

func (d *DocumentQueryCustomizationDelegate) waitForNonStaleResults() *IDocumentQueryCustomization {
	//query._waitForNonStaleResults(null);
	return d
}

func (d *DocumentQueryCustomizationDelegate) waitForNonStaleResultsWithTimeout(waitTimeout time.Duration) *IDocumentQueryCustomization {
	//query._waitForNonStaleResults(waitTimeout);
	return d
}
