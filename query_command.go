package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &QueryCommand{}
)

type QueryCommand struct {
	RavenCommandBase

	conventions      *DocumentConventions
	indexQuery       *IndexQuery
	metadataOnly     bool
	indexEntriesOnly bool

	Result *QueryResult
}

func NewQueryCommand(conventions *DocumentConventions, indexQuery *IndexQuery, metadataOnly bool, indexEntriesOnly bool) (*QueryCommand, error) {
	if indexQuery == nil {
		return nil, newIllegalArgumentError("IndexQuery cannot be null")
	}
	cmd := &QueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions:      conventions,
		indexQuery:       indexQuery,
		metadataOnly:     metadataOnly,
		indexEntriesOnly: indexEntriesOnly,
	}
	cmd.IsReadRequest = true
	return cmd, nil
}

func (c *QueryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	c.CanCache = !c.indexQuery.disableCaching

	// we won't allow aggressive caching of queries with WaitForNonStaleResults
	c.CanCacheAggressively = c.CanCache && !c.indexQuery.waitForNonStaleResults

	// we need to add a query hash because we are using POST queries
	// so we need to unique parameter per query so the query cache will
	// work properly
	path := node.URL + "/databases/" + node.Database + "/queries?queryHash=" + c.indexQuery.GetQueryHash()

	if c.metadataOnly {
		path += "&metadataOnly=true"
	}

	if c.indexEntriesOnly {
		path += "&debug=entries"
	}

	m := jsonExtensionsWriteIndexQuery(c.conventions, c.indexQuery)
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
