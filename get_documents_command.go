package ravendb

import (
	"encoding/json"
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

		_includes:     includes,
		_metadataOnly: metadataOnly,
		_start:        -1,
		_pageSize:     -1,
	}

	if len(ids) == 1 {
		cmd._id = ids[0]
	} else {
		cmd._ids = ids
	}
	cmd.IsReadRequest = true
	return cmd
}

func NewGetDocumentsCommandFull(startWith string, startAfter string, matches string, exclude string, start int, pageSize int, metadataOnly bool) *GetDocumentsCommand {
	panicIf(startWith == "", "startWith cannot be null")
	return &GetDocumentsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_startWith:    startWith,
		_startAfter:   startAfter,
		_matches:      matches,
		_exclude:      exclude,
		_start:        start,
		_pageSize:     pageSize,
		_metadataOnly: metadataOnly,
	}
}

func (c *GetDocumentsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?"
	if c._start != -1 {
		url += "&start=" + strconv.Itoa(c._start)
	}

	if c._pageSize != -1 {
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
	} else if len(c._ids) > 0 {
		return c.prepareRequestWithMultipleIds(url)
	}

	return NewHttpGet(url)
}

func (c *GetDocumentsCommand) prepareRequestWithMultipleIds(url string) (*http.Request, error) {
	ids := c._ids
	uniqueIds := NewSet_String()
	for _, id := range ids {
		uniqueIds.add(id)
	}
	totalLen := 0
	for _, s := range uniqueIds.strings {
		totalLen += len(s)
	}

	// if it is too big, we drop to POST (note that means that we can't use the HTTP cache any longer)
	// we are fine with that, requests to load > 1024 items are going to be rare
	isGet := totalLen < 1024

	if isGet {
		for _, s := range uniqueIds.strings {
			url += "&id=" + UrlUtils_escapeDataString(s)
		}
		return NewHttpGet(url)
	}

	m := map[string]interface{}{
		"Ids": uniqueIds.strings,
	}
	d, err := json.Marshal(m)
	panicIf(err != nil, "json.Marshal() failed with %s", err)
	return NewHttpPost(url, string(d))
}

func (c *GetDocumentsCommand) setResponse(response string, fromCache bool) error {
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
