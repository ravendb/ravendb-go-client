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

func (r *ExplainQueryResult) getIndex() string {
	return r.Index
}

func (r *ExplainQueryResult) setIndex(index string) {
	r.Index = index
}

func (r *ExplainQueryResult) getReason() string {
	return r.Reason
}

func (r *ExplainQueryResult) setReason(reason string) {
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

func (c *ExplainQueryCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/queries?debug=explain"
	// TODO:
	/*
	   request.setEntity(new ContentProviderHttpEntity(outputStream -> {
	       try (JsonGenerator generator = mapper.getFactory().createGenerator(outputStream)) {
	           JsonExtensions.writeIndexQuery(generator, _conventions, _indexQuery);
	       } catch (IOException e) {
	           throw new RuntimeException(e);
	       }
	   }, ContentType.APPLICATION_JSON));
	*/
	return NewHttpPost(url, nil)
}

func (c *ExplainQueryCommand) setResponse(response []byte, fromCache bool) error {
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
