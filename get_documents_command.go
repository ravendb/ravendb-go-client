package ravendb

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

var (
	_ RavenCommand = &GetDocumentsCommand{}
)

type GetDocumentsCommand struct {
	*RavenCommandBase

	_id string

	_ids      []string
	_includes []string

	_metadataOnly bool

	_startWith  string
	_matches    string
	_start      int
	_pageSize   int
	_exclude    string
	_startAfter string

	Result *GetDocumentsResult
}

func NewGetDocumentsCommand(ids []string, includes []string, metadataOnly bool) *GetDocumentsCommand {
	cmd := &GetDocumentsCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_includes:        includes,
		_metadataOnly:    metadataOnly,
	}

	if len(ids) == 1 {
		cmd._id = ids[0]
	} else {
		cmd._ids = ids
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetDocumentsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?"
	// TODO: is _start == 0 valid?
	if c._start != 0 {
		url += "&start=" + strconv.Itoa(c._start)
	}

	if c._pageSize != 0 {
		url += "&pageSize=" + strconv.Itoa(c._pageSize)
	}

	if c._metadataOnly {
		url += "&metadataOnly=true"
	}

	if c._startWith != "" {
		url += "&startsWith="
		url += UrlUtils_escapeDataString(c._startWith)

		if c._matches != "" {
			url += "&matches="
			url += c._matches
		}

		if c._exclude != "" {
			url += "&exclude="
			url += c._exclude
		}

		if c._startAfter != "" {
			url += "&startAfter="
			url += c._startAfter
		}
	}

	for _, include := range c._includes {
		url += "&include="
		url += include
	}

	if c._id != "" {
		url += "&id="
		url += UrlUtils_escapeDataString(c._id)
		return NewHttpGet(url)
	}

	panicIf(len(c._ids) == 0, "must provide _id or _ids")

	return c.prepareRequestWithMultipleIds(url)
}

func (c *GetDocumentsCommand) prepareRequestWithMultipleIds(url string) (*http.Request, error) {
	//ids := c._ids
	panicIf(true, "NYI")
	return nil, errors.New("NYI")
}

func (c *GetDocumentsCommand) setResponse(response String, fromCache bool) error {
	if response == "" {
		return nil
	}

	var res GetDocumentsResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
