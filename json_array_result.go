package ravendb

// ArrayNode represents result of BatchCommand, which is array of JSON objects
// it's a type alias so that it doesn't need casting when json marshalling
type ArrayNode = []ObjectNode

// JSONArrayResult describes server's JSON response to batch command
type JSONArrayResult struct {
	Results ArrayNode `json:"Results"`
}
