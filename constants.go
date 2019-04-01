package ravendb

const (
	// Name of struct field that represents identity property
	IdentityProperty = "ID"

	MetadataCollection             = "@collection"
	MetadataProjection             = "@projection"
	MetadataKey                    = "@metadata"
	MetadataID                     = "@id"
	MetadataConflict               = "@conflict"
	MetadataIDProperty             = "Id"
	MetadataFlags                  = "@flags"
	MetadataAttachments            = "@attachments"
	COUNTERS                       = "@counters"
	REVISION_COUNTERS              = "@counters-snapshot"
	MetadataInddexScore            = "@index-score"
	MetadataLastModified           = "@last-modified"
	MetadataRavenGoType            = "Raven-Go-Type"
	MetadataChangeVector           = "@change-vector"
	MetadataExpires                = "@expires"
	MetadataAllDocumentsCollection = "@all_docs"

	IndexingSideBySideIndexNamePrefix = "ReplacementOf/"
	IndexingFieldNameDocumentID       = "id()"
	IndexingFieldNameReduceKeyHash    = "hash(key())"
	IndexingFieldNameReduceKeyValue   = "key()"
	IndexingFieldAllFields            = "__all_fields"
	IndexingFieldsNameSpatialShare    = "spatial(shape)"
	//TBD CUSTOM_SORT_FIELD_NAME = "__customSort";
	IndexingSpatialDefaultDistnaceErrorPct = 0.025

	headersRequestTime                   = "Raven-Request-Time"
	headersRefreshTopology               = "Refresh-Topology"
	headersTopologyEtag                  = "Topology-Etag"
	LAST_KNOWN_CLUSTER_TRANSACTION_INDEX = "Known-Raft-Index"
	headersClientConfigurationEtag       = "Client-Configuration-Etag"
	headersRefreshClientConfiguration    = "Refresh-Client-Configuration"
	headersClientVersion                 = "Raven-Client-Version"
	SERVER_VERSION                       = "Raven-Server-Version"
	headersEtag                          = "ETag"
	headersIfNoneMatch                   = "If-None-Match"

	Counters_All = "@all_counters"
)
