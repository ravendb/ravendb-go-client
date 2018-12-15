package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &QueryCommand{}
)

type QueryCommand struct {
	RavenCommandBase

	_conventions      *DocumentConventions
	_indexQuery       *IndexQuery
	_metadataOnly     bool
	_indexEntriesOnly bool

	Result *QueryResult
}

func NewQueryCommand(conventions *DocumentConventions, indexQuery *IndexQuery, metadataOnly bool, indexEntriesOnly bool) *QueryCommand {
	panicIf(indexQuery == nil, "IndexQuery cannot be null")
	cmd := &QueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:      conventions,
		_indexQuery:       indexQuery,
		_metadataOnly:     metadataOnly,
		_indexEntriesOnly: indexEntriesOnly,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *QueryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	c.CanCache = !c._indexQuery.disableCaching

	// we won't allow aggressive caching of queries with WaitForNonStaleResults
	c.CanCacheAggressively = c.CanCache && !c._indexQuery.waitForNonStaleResults

	// we need to add a query hash because we are using POST queries
	// so we need to unique parameter per query so the query cache will
	// work properly
	path := node.GetUrl() + "/databases/" + node.GetDatabase() + "/queries?queryHash=" + c._indexQuery.GetQueryHash()

	if c._metadataOnly {
		path += "&metadataOnly=true"
	}

	if c._indexEntriesOnly {
		path += "&debug=entries"
	}

	m := JsonExtensions_writeIndexQuery(c._conventions, c._indexQuery)
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(path, d)
}

func (c *QueryCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
