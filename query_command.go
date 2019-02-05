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

func NewQueryCommand(conventions *DocumentConventions, indexQuery *IndexQuery, metadataOnly bool, indexEntriesOnly bool) (*QueryCommand, error) {
	if indexQuery == nil {
		return nil, newIllegalArgumentError("IndexQuery cannot be null")
	}
	cmd := &QueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:      conventions,
		_indexQuery:       indexQuery,
		_metadataOnly:     metadataOnly,
		_indexEntriesOnly: indexEntriesOnly,
	}
	cmd.IsReadRequest = true
	return cmd, nil
}

func (c *QueryCommand) createRequest(node *ServerNode) (*http.Request, error) {
	c.CanCache = !c._indexQuery.disableCaching

	// we won't allow aggressive caching of queries with WaitForNonStaleResults
	c.CanCacheAggressively = c.CanCache && !c._indexQuery.waitForNonStaleResults

	// we need to add a query hash because we are using POST queries
	// so we need to unique parameter per query so the query cache will
	// work properly
	path := node.URL + "/databases/" + node.Database + "/queries?queryHash=" + c._indexQuery.GetQueryHash()

	if c._metadataOnly {
		path += "&metadataOnly=true"
	}

	if c._indexEntriesOnly {
		path += "&debug=entries"
	}

	m := jsonExtensionsWriteIndexQuery(c._conventions, c._indexQuery)
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return newHttpPost(path, d)
}

func (c *QueryCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
