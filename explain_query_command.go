package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &ExplainQueryCommand{}
)

type ExplainQueryResult struct {
	Index  string `json:"Index"`
	Reason string `json:"Reason"`
}

type ExplainQueryCommand struct {
	RavenCommandBase

	_conventions *DocumentConventions
	_indexQuery  *IndexQuery

	Result []*ExplainQueryResult
}

func NewExplainQueryCommand(conventions *DocumentConventions, indexQuery *IndexQuery) *ExplainQueryCommand {
	panicIf(conventions == nil, "Conventions cannot be null")
	panicIf(indexQuery == nil, "IndexQuery cannot be null")
	cmd := &ExplainQueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions: conventions,
		_indexQuery:  indexQuery,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *ExplainQueryCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/queries?debug=explain"

	v := jsonExtensionsWriteIndexQuery(c._conventions, c._indexQuery)
	d, err := jsonMarshal(v)
	panicIf(err != nil, "jsonMarshal() failed with %s", err)
	return NewHttpPost(url, d)
}

func (c *ExplainQueryCommand) setResponse(response []byte, fromCache bool) error {
	var res struct {
		Results []*ExplainQueryResult
	}
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	if res.Results == nil {
		return throwInvalidResponse()
	}
	c.Result = res.Results
	return nil
}
