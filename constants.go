package ravendb

const (
	Constants_Documents_Metadata_COLLECTION      = "@collection"
	Constants_Documents_Metadata_PROJECTION      = "@projection"
	Constants_Documents_Metadata_KEY             = "@metadata"
	Constants_Documents_Metadata_ID              = "@id"
	Constants_Documents_Metadata_CONFLICT        = "@conflict"
	Constants_Documents_Metadata_ID_PROPERTY     = "Id"
	Constants_Documents_Metadata_FLAGS           = "@flags"
	Constants_Documents_Metadata_ATTACHMENTS     = "@attachments"
	Constants_Documents_Metadata_INDEX_SCORE     = "@index-score"
	Constants_Documents_Metadata_LAST_MODIFIED   = "@last-modified"
	Constants_Documents_Metadata_RAVEN_JAVA_TYPE = "Raven-Java-Type"
	Constants_Documents_Metadata_CHANGE_VECTOR   = "@change-vector"
	Constants_Documents_Metadata_EXPIRES         = "@expires"

	Constants_Documents_Indexing_SIDE_BY_SIDE_INDEX_NAME_PREFIX = "ReplacementOf/"

	Constants_Documents_Indexing_Fields_DOCUMENT_ID_FIELD_NAME          = "id()"
	Constants_Documents_Indexing_Fields_REDUCE_KEY_HASH_FIELD_NAME      = "hash(key())"
	Constants_Documents_Indexing_Fields_REDUCE_KEY_KEY_VALUE_FIELD_NAME = "key()"
	Constants_Documents_Indexing_Fields_ALL_FIELDS                      = "__all_fields"
	Constants_Documents_Indexing_Fields_SPATIAL_SHAPE_FIELD_NAME        = "spatial(shape)"
	//TBD CUSTOM_SORT_FIELD_NAME = "__customSort";

	Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT = 0.025

	Constants_Headers_REQUEST_TIME = "Raven-Request-Time"

	Constants_Headers_REFRESH_TOPOLOGY = "Refresh-Topology"

	Constants_Headers_TOPOLOGY_ETAG = "Topology-Etag"

	Constants_Headers_CLIENT_CONFIGURATION_ETAG = "Client-Configuration-Etag"

	Constants_Headers_REFRESH_CLIENT_CONFIGURATION = "Refresh-Client-Configuration"

	Constants_Headers_ETAG = "ETag"

	Constants_Headers_IF_NONE_MATCH     = "If-None-Match"
	Constants_Headers_TRANSFER_ENCODING = "Transfer-Encoding"
	Constants_Headers_CONTENT_ENCODING  = "Content-Encoding"
	Constants_Headers_CONTENT_LENGTH    = "Content-Length"
)
