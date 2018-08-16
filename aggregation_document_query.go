package ravendb

type IAggregationDocumentQuery = AggregationDocumentQuery

// Note: AggregationDocumentQuery is fused into AggregationQueryBase because
// in Java AggregationQueryBase calls functions implemented in AggregationDocumentQuery
// and that doesn't translate to Go's embedding
type AggregationDocumentQuery = AggregationQueryBase
