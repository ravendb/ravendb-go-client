package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &ExplainQueryCommand{}
)

type ExplainQueryResult struct {
	Index  string `json:"Index"`
	Reason string `json:"Reason"`
}

func (r *ExplainQueryResult) GetIndex() string {
	return r.Index
}

func (r *ExplainQueryResult) SetIndex(index string) {
	r.Index = index
}

func (r *ExplainQueryResult) GetReason() string {
	return r.Reason
}

func (r *ExplainQueryResult) SetReason(reason string) {
	r.Reason = reason
}

type ExplainQueryCommand struct {
	*RavenCommandBase

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

func (c *ExplainQueryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/queries?debug=explain"

	v := JsonExtensions_writeIndexQuery(c._conventions, c._indexQuery)
	d, err := json.Marshal(v)
	panicIf(err != nil, "json.Marshal() failed with %s", err)
	return NewHttpPost(url, d)
}

func (c *ExplainQueryCommand) SetResponse(response []byte, fromCache bool) error {
	var res struct {
		Results []*ExplainQueryResult
	}
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	if res.Results == nil {
		return throwInvalidResponse()
	}
	c.Result = res.Results
	return nil
}
